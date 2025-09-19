import { inject } from '@angular/core';
import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { map, Observable } from 'rxjs';
import { AuthService } from './auth.service';

/**
 * This guard function checks if a user is authenticated before allowing access to a route.
 */
export const authGuard: CanActivateFn = (route, state): Observable<boolean | UrlTree> => {
  
  // Inject the AuthService and Router
  const authService = inject(AuthService);
  const router = inject(Router);

  // Use the isAuthenticated$ observable from the AuthService
  return authService.isAuthenticated$.pipe(
    map(isAuthenticated => {
      if (isAuthenticated) {
        // If the user is authenticated, allow access.
        return true;
      } else {
        // If the user is not authenticated, redirect them to the login page.
        // Returning a UrlTree tells the router to navigate somewhere else.
        console.log('Access denied. Redirecting to login...');
        return router.createUrlTree(['/login']);
      }
    })
  );
};