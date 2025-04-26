import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { provideAppInitializer, inject } from '@angular/core';
import { AppComponent }  from './app/app.component';
import { appRoutes } from './app/app.routes';
import { AuthInterceptor } from './app/core/auth.interceptor';
import { AuthService } from './app/core/auth.service';

bootstrapApplication(AppComponent, {
  providers: [
    provideRouter(appRoutes),

    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true
    },

    provideHttpClient(withInterceptorsFromDi()),

    provideAppInitializer(() => inject(AuthService).init())
  ],
}).catch(err => console.error(err));
