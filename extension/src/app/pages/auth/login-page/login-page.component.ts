import { Component } from '@angular/core';
import { AuthService } from 'core/auth.service';

@Component({
  selector: 'login-page',
  standalone: true,
  imports: [],
  templateUrl: './login-page.component.html',
  styleUrl: './login-page.component.scss'
})
export class LoginPageComponent {
  constructor(private auth: AuthService) {}

  onLogin() {
    this.auth.login().subscribe({
      next: () => {
        // after login, angular router will take you to dashboard via your routes
      },
      error: err => console.error('Login failed', err)
    });
  }
}