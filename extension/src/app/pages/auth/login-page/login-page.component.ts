import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { take } from 'rxjs';
import { AuthService } from 'core/auth.service';

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './login-page.component.html',
  styleUrls: ['./login-page.component.scss']
})
export class LoginPageComponent implements OnInit {
  constructor(
    private auth: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    // If already logged in, go straight to dashboard
    this.auth.isLoggedIn$
      .pipe(take(1))
      .subscribe(loggedIn => {
        if (loggedIn) {
          this.router.navigate(['dashboard']);
        }
      });
  }

  onLogin(): void {
    this.auth.login().subscribe({
      next: () => {
        // After a successful login, route to dashboard
        this.router.navigate(['dashboard']);
      },
      error: err => console.error('Login failed', err)
    });
  }
}