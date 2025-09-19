import { HttpInterceptorFn, HttpRequest, HttpHandlerFn, HttpEvent } from '@angular/common/http';
import { inject } from '@angular/core';
import { from, Observable, switchMap } from 'rxjs';
import { AuthService } from './auth.service';

/**
 * This interceptor function automatically attaches the JWT bearer token
 * to outgoing HTTP requests to your API.
 */
export const authInterceptor: HttpInterceptorFn = (
  req: HttpRequest<unknown>,
  next: HttpHandlerFn
): Observable<HttpEvent<unknown>> => {
  
  const authService = inject(AuthService);

  // The getToken() method in our AuthService returns a Promise.
  // We convert it to an Observable to use it in the RxJS pipe.
  return from(authService.getToken()).pipe(
    switchMap(token => {
      // If no token exists, pass the original request along without changes.
      if (!token) {
        return next(req);
      }

      // If a token exists, clone the request and add the Authorization header.
      const authReq = req.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`
        }
      });
      
      // Pass the cloned, authorized request to the next handler.
      return next(authReq);
    })
  );
};