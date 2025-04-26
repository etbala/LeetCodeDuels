import { Injectable } from '@angular/core';
import { BehaviorSubject, from, switchMap, tap } from 'rxjs';
import { chromeIdentityLaunchFlow, chromeStorageGet, chromeStorageSet } from './chrome.utils';
import { HttpClient } from '@angular/common/http';

interface JwtPayload {
  username: string;
  // …any other fields your backend includes
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private tokenKey = 'token';
  private userKey = 'github_username';

  private _isLoggedIn$ = new BehaviorSubject<boolean>(false);
  isLoggedIn$ = this._isLoggedIn$.asObservable();

  private _username$ = new BehaviorSubject<string|null>(null);
  username$ = this._username$.asObservable();

  constructor(private http: HttpClient) {
    // On service init, try to load existing token
    from(chromeStorageGet<string>(this.tokenKey))
      .pipe(
        tap(token => {
          if (token) this.setSession(token);
        })
      )
      .subscribe();
  }

  /** Kick off GitHub OAuth flow, exchange code, persist JWT */
  login() {
    const clientId = 'Ov23liQ4ERGhUYdeT8yb';
    const redirectUri = chrome.identity.getRedirectURL();
    const authURL = `https://github.com/login/oauth/authorize`
      + `?client_id=${encodeURIComponent(clientId)}`
      + `&redirect_uri=${encodeURIComponent(redirectUri)}`
      + `&scope=user`;

    return from(chromeIdentityLaunchFlow(authURL)).pipe(
      switchMap(redirectUrl => {
        const code = new URL(redirectUrl).searchParams.get('code');
        if (!code) throw new Error('No code returned');
        return this.http.post<{ token: string }>(
          'http://localhost:8080/oauth/exchange-token',
          { code }
        );
      }),
      tap(({ token }) => this.setSession(token))
    );
  }

  /** Clears storage and resets state */
  logout(): Promise<void> {
    this._isLoggedIn$.next(false);
    this._username$.next(null);
    return chromeStorageSet(this.tokenKey, '').then(() => 
      chromeStorageSet(this.userKey, '')
    );
  }

  private setSession(token: string) {
    // save token
    chromeStorageSet(this.tokenKey, token);
    // decode payload
    const payload: JwtPayload = JSON.parse(atob(token.split('.')[1]));
    this._username$.next(payload.username);
    this._isLoggedIn$.next(true);
    // persist username for quick access
    chromeStorageSet(this.userKey, payload.username);
  }

  /** Sync-ready getter for token if you need it */
  get token(): Promise<string|null> {
    return chromeStorageGet(this.tokenKey);
  }
}
