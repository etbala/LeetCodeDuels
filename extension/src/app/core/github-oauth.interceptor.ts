import { HttpInterceptorFn } from '@angular/common/http';

export const githubOauthInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req);
};
