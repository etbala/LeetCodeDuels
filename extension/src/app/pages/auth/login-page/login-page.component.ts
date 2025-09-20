import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { AuthService } from 'app/services/auth.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-login-page',
  imports: [CommonModule],
  templateUrl: './login-page.component.html',
  styleUrls: ['./login-page.component.scss']
})
export class LoginPageComponent implements OnInit {
  isLoading = false;
  error: string | null = null;
  returnUrl: string = '/';

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit() {
    this.returnUrl = this.route.snapshot.queryParams['returnUrl'] || '/';
    
    this.authService.isAuthenticated().then(isAuth => {
      if (isAuth) {
        this.router.navigate([this.returnUrl]);
      }
    });
  }

  async loginWithGitHub() {
    this.isLoading = true;
    this.error = null;
    
    try {
      await this.authService.login();
      this.router.navigate([this.returnUrl]);
    } catch (error: any) {
      console.error('Login failed:', error);
      this.error = error.message || 'An error occurred during authentication. Please try again.';
    } finally {
      this.isLoading = false;
    }
  }
}