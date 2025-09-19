import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../services/auth.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.scss'],
})
export class HeaderComponent {
  constructor(
    private auth: AuthService,
    private router: Router
  ) {}

  logout(): void {
    this.auth.logout().then(() => {
      // After clearing the JWT, redirect to login
      this.router.navigate(['/login']);
    });
  }
}