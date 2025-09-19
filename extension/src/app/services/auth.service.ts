import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, lastValueFrom } from 'rxjs';

// This tells TypeScript that the 'chrome' object exists, as it's injected by the browser extension environment.
declare var chrome: any;

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  // Use a BehaviorSubject to hold and stream the authentication state.
  private authState = new BehaviorSubject<boolean>(false);
  
  // Expose the auth state as a public observable for components to subscribe to.
  public isAuthenticated$ = this.authState.asObservable();

  // Your backend and GitHub App configuration
  private readonly GITHUB_CLIENT_ID = 'Ov23liQ4ERGhUYdeT8yb';
  private readonly API_BASE_URL = 'http://localhost:8080';

  constructor(private http: HttpClient) {
    // When the service is initialized, check if the user is already logged in.
    this.checkInitialAuthState();
  }

  /**
   * Checks for an existing token in extension storage to set the initial auth state.
   */
  private async checkInitialAuthState(): Promise<void> {
    const token = await this.getToken();
    if (token) {
      // You might want to add a step here to verify the token with your backend.
      this.authState.next(true);
    }
  }

  /**
   * The main login function that orchestrates the entire OAuth flow.
   */
  async login(): Promise<void> {
    try {
      // 1. Construct the GitHub authorization URL
      const redirectUri = `${this.API_BASE_URL}/auth/github/callback`;
      const authUrl = `https://github.com/login/oauth/authorize?client_id=${this.GITHUB_CLIENT_ID}&redirect_uri=${encodeURIComponent(redirectUri)}&scope=read:user`;

      // 2. Launch the web auth flow using the extension's identity API
      const redirectUrl = await new Promise<string>((resolve, reject) => {
        chrome.identity.launchWebAuthFlow({ url: authUrl, interactive: true }, (responseUrl?: string) => {
          if (chrome.runtime.lastError || !responseUrl) {
            reject(new Error(chrome.runtime.lastError?.message || 'The user cancelled the web auth flow.'));
          } else {
            resolve(responseUrl);
          }
        });
      });
      
      // 3. Extract the authorization code from the URL
      const code = new URL(redirectUrl).searchParams.get('code');
      if (!code) {
        throw new Error('Authorization code not found in redirect URL.');
      }

      // 4. Exchange the code for a JWT from your backend using HttpClient
      const tokenResponse$ = this.http.post<{ token: string }>(`${this.API_BASE_URL}/auth/github/exchange`, { code });
      const tokenResponse = await lastValueFrom(tokenResponse$);
      
      // 5. Save the token and update the auth state
      await this.saveToken(tokenResponse.token);
      this.authState.next(true);
      console.log('Login successful!');

    } catch (error) {
      console.error('Login failed:', error);
      this.authState.next(false); // Ensure state is false on failure
    }
  }

  /**
   * Logs the user out by removing the token and updating the state.
   */
  async logout(): Promise<void> {
    await new Promise<void>((resolve) => {
      chrome.storage.local.remove('authToken', () => resolve());
    });
    this.authState.next(false);
  }

  /**
   * Retrieves the auth token from the browser extension's local storage.
   */
  public getToken(): Promise<string | null> {
    return new Promise((resolve) => {
      chrome.storage.local.get('authToken', (result: { [key: string]: any }) => {
        // FIX IS HERE: Use bracket notation for properties from an index signature.
        resolve(result['authToken'] || null);
      });
    });
  }

  /**
   * Saves the auth token to the browser extension's local storage.
   */
  private saveToken(token: string): Promise<void> {
    return new Promise((resolve) => {
      chrome.storage.local.set({ authToken: token }, () => resolve());
    });
  }
}
