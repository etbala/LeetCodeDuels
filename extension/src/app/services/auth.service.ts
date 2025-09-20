import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject } from 'rxjs';
import { environment } from '../../environments/environment.prod';

interface User {
  id: string;
  username: string;
  lc_username: string;
  avatar_url: string;
}

interface AuthResponse {
  token: string;
  user: User;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly API_URL = environment.apiUrl;
  private readonly GITHUB_CLIENT_ID = environment.githubClientId;
  private readonly STORAGE_KEY = 'auth_token';
  private readonly USER_KEY = 'user_data';
  
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();
  
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();
  
  private authWindow: Window | null = null;

  constructor(private http: HttpClient) {
    this.initializeAuth();
  }

  private async initializeAuth(): Promise<void> {
    try {
      const token = await this.getStoredToken();
      const user = await this.getStoredUser();
      
      if (token && user) {
        this.currentUserSubject.next(user);
        this.isAuthenticatedSubject.next(true);
      }
    } catch (error) {
      console.error('Error initializing auth:', error);
      this.clearAuth();
    }
  }

  // Initiates GitHub OAuth login flow
  async login(): Promise<void> {
    return new Promise((resolve, reject) => {
      const state = this.generateState();
      sessionStorage.setItem('oauth_state', state);
      
      const params = new URLSearchParams({
        client_id: this.GITHUB_CLIENT_ID,
        redirect_uri: `${this.API_URL}/auth/github/callback`,
        scope: 'user:email',
        state: state
      });
      
      const authUrl = `https://github.com/login/oauth/authorize?${params.toString()}`;
      
      this.authWindow = window.open(
        authUrl,
        'github-auth',
        'width=600,height=700,left=100,top=100'
      );
      
      if (!this.authWindow) {
        reject(new Error('Failed to open authentication window. Please check popup blocker settings.'));
        return;
      }
      
      // Listen for OAuth callback message
      const messageListener = async (event: MessageEvent) => {
        if (!event.origin.startsWith(this.API_URL.replace(/\/+$/, ''))) {
          return;
        }
        
        // Check if it's our OAuth callback
        if (event.data?.type === 'github-oauth-callback') {
          window.removeEventListener('message', messageListener);
          
          if (this.authWindow && !this.authWindow.closed) {
            this.authWindow.close();
          }
          
          if (event.data.error) {
            reject(new Error(`Authentication failed: ${event.data.error}`));
            return;
          }
          
          if (event.data.code) {
            try {
              await this.exchangeCodeForToken(event.data.code);
              resolve();
            } catch (error) {
              reject(error);
            }
          } else {
            reject(new Error('No authorization code received'));
          }
        }
      };
      
      window.addEventListener('message', messageListener);
      
      // Clean up if window is closed without completing auth
      const windowChecker = setInterval(() => {
        if (this.authWindow && this.authWindow.closed) {
          clearInterval(windowChecker);
          window.removeEventListener('message', messageListener);
          reject(new Error('Authentication window was closed'));
        }
      }, 1000);
      
      // Timeout after 5 minutes
      setTimeout(() => {
        clearInterval(windowChecker);
        window.removeEventListener('message', messageListener);
        if (this.authWindow && !this.authWindow.closed) {
          this.authWindow.close();
        }
        reject(new Error('Authentication timeout'));
      }, 5 * 60 * 1000);
    });
  }

  // Exchange authorization code for JWT token
  private async exchangeCodeForToken(code: string): Promise<void> {
    try {
      const response = await this.http.post<AuthResponse>(
        `${this.API_URL}/auth/github/exchange`,
        { code }
      ).toPromise();
      
      if (!response || !response.token) {
        throw new Error('Invalid response from server');
      }
      
      await this.setStoredToken(response.token);
      await this.setStoredUser(response.user);
      
      this.currentUserSubject.next(response.user);
      this.isAuthenticatedSubject.next(true);
    } catch (error) {
      console.error('Token exchange failed:', error);
      throw error;
    }
  }

  // Logout user and clear stored data
  async logout(): Promise<void> {
    await this.clearAuth();
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);
  }

  // Get stored authentication token
  async getToken(): Promise<string | null> {
    return this.getStoredToken();
  }

  // Check if user is currently authenticated
  async isAuthenticated(): Promise<boolean> {
    const token = await this.getStoredToken();
    return !!token;
  }

  // Get current user
  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  private async getStoredToken(): Promise<string | null> {
    if (typeof chrome !== 'undefined' && chrome.storage) {
      return new Promise((resolve) => {
        chrome.storage.local.get([this.STORAGE_KEY], (result) => {
          resolve(result[this.STORAGE_KEY] || null);
        });
      });
    }
    // Fallback to localStorage for development
    return localStorage.getItem(this.STORAGE_KEY);
  }

  private async setStoredToken(token: string): Promise<void> {
    if (typeof chrome !== 'undefined' && chrome.storage) {
      return new Promise((resolve) => {
        chrome.storage.local.set({ [this.STORAGE_KEY]: token }, () => {
          resolve();
        });
      });
    }
    // Fallback to localStorage for development
    localStorage.setItem(this.STORAGE_KEY, token);
  }

  private async getStoredUser(): Promise<User | null> {
    if (typeof chrome !== 'undefined' && chrome.storage) {
      return new Promise((resolve) => {
        chrome.storage.local.get([this.USER_KEY], (result) => {
          const userData = result[this.USER_KEY];
          resolve(userData ? JSON.parse(userData) : null);
        });
      });
    }
    // Fallback to localStorage for development
    const userData = localStorage.getItem(this.USER_KEY);
    return userData ? JSON.parse(userData) : null;
  }

  private async setStoredUser(user: User): Promise<void> {
    const userString = JSON.stringify(user);
    if (typeof chrome !== 'undefined' && chrome.storage) {
      return new Promise((resolve) => {
        chrome.storage.local.set({ [this.USER_KEY]: userString }, () => {
          resolve();
        });
      });
    }
    // Fallback to localStorage for development
    localStorage.setItem(this.USER_KEY, userString);
  }

  private async clearAuth(): Promise<void> {
    if (typeof chrome !== 'undefined' && chrome.storage) {
      return new Promise((resolve) => {
        chrome.storage.local.remove([this.STORAGE_KEY, this.USER_KEY], () => {
          resolve();
        });
      });
    }
    // Fallback to localStorage for development
    localStorage.removeItem(this.STORAGE_KEY);
    localStorage.removeItem(this.USER_KEY);
  }

  private generateState(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  }
}