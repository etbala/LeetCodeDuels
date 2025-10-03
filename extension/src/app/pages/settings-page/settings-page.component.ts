import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-settings-page',
  imports: [CommonModule],
  templateUrl: './settings-page.component.html',
  styleUrl: './settings-page.component.scss'
})
export class SettingsPageComponent {
  constructor(
    private auth: AuthService,
    private router: Router
  ) {}

  async logout() {
    this.auth.logout().then(() => {
      // After clearing the JWT, redirect to login
      this.router.navigate(['/login']);
    });
  }
}
