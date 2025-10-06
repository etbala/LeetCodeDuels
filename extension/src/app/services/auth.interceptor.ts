import { inject } from '@angular/core';
import { HttpInterceptorFn, HttpRequest, HttpHandlerFn, HttpEvent, HttpErrorResponse } from '@angular/common/http';
import { Observable, from, throwError } from 'rxjs';
import { switchMap, catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req: HttpRequest<unknown>, next: HttpHandlerFn): Observable<HttpEvent<unknown>> => {
  const authService = inject(AuthService);
  const router = inject(Router);

  // todo: only add to urls starting with server base url
  if (req.url.includes('/auth/')) {
    return next(req);
  }

  return from(authService.getToken()).pipe(
    switchMap(token => {
      let authReq = req;

      // If a token exists, clone the request to add the new header.
      if (token) {
        authReq = req.clone({
          setHeaders: {
            Authorization: `Bearer ${token}`
          }
        });
      }

      // Pass the cloned or original request to the next handler
      return next(authReq).pipe(
        catchError((error: HttpErrorResponse) => {
          // If a 401 error occurs, the token is invalid/expired.
          if (error.status === 401) {
            authService.logout().then(() => {
              router.navigate(['/login']);
            });
          }
          // Re-throw the error to be handled by the calling service.
          return throwError(() => error);
        })
      );
    })
  );
};