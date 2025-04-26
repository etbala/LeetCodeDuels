import {
    HttpEvent, HttpHandler, HttpInterceptor, HttpRequest
} from '@angular/common/http';
import { Injectable } from '@angular/core';
import { from, Observable } from 'rxjs';
import { switchMap } from 'rxjs/operators';
import { AuthService } from './auth.service';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
constructor(private auth: AuthService) {}

intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return from(this.auth.token).pipe(
    switchMap(token => {
        if (!token) return next.handle(req);
        const cloned = req.clone({
        setHeaders: { Authorization: `Bearer ${token}` }
        });
        return next.handle(cloned);
    })
    );
}
}