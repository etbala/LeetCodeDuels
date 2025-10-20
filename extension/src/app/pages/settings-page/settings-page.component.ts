import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { finalize } from 'rxjs/operators';
import { AuthService } from '../../services/auth/auth.service';
import { UserService } from '../../services/user/user.service';
import { User } from 'app/models/user.model';

@Component({
  selector: 'app-settings-page',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './settings-page.component.html',
  styleUrl: './settings-page.component.scss'
})
export class SettingsPageComponent implements OnInit {
  usernameControl: FormControl;
  lcUsernameControl: FormControl;

  currentUser: User | null = null;
  
  isLoading = true;
  isEditingUsername = false;
  isEditingLcUsername = false;
  isSavingUsername = false;
  isSavingLcUsername = false;
  
  message: { type: 'success' | 'error', text: string } | null = null;

  constructor(
    private auth: AuthService,
    private router: Router,
    private userService: UserService
  ) {
    this.usernameControl = new FormControl('', [Validators.required, Validators.maxLength(32)]);
    this.lcUsernameControl = new FormControl('', [Validators.maxLength(50)]);
  }

  ngOnInit(): void {
    this.userService.getMyProfile()
      .pipe(finalize(() => this.isLoading = false))
      .subscribe({
        next: (user) => {
          this.currentUser = user;
          this.usernameControl.setValue(user.username);
          this.lcUsernameControl.setValue(user.lc_username);
        },
        error: (err) => {
          console.error("Failed to load user profile:", err);
          this.message = { type: 'error', text: 'Could not load your profile. Please refresh.' };
        }
      });
  }

  editUsername(): void {
    this.isEditingUsername = true;
    this.message = null;
  }

  cancelEditUsername(): void {
    this.isEditingUsername = false;
    this.usernameControl.setValue(this.currentUser?.username);
  }

  saveUsername(): void {
    if (this.usernameControl.invalid || this.isSavingUsername) {
      return;
    }
    
    this.isSavingUsername = true;
    this.message = null;
    
    this.userService.updateUser({ username: this.usernameControl.value })
      .pipe(finalize(() => this.isSavingUsername = false))
      .subscribe({
        next: (updatedUser) => {
          if (this.currentUser) {
            this.currentUser.username = updatedUser.username;
            this.currentUser.discriminator = updatedUser.discriminator;
          }
          this.isEditingUsername = false;
          this.message = { type: 'success', text: 'Display name updated!' };
        },
        error: (err) => {
          console.error("Failed to update username:", err);
          this.message = { type: 'error', text: 'Failed to update. Display name may be taken.' };
        }
      });
  }

  editLcUsername(): void {
    this.isEditingLcUsername = true;
    this.message = null;
  }

  cancelEditLcUsername(): void {
    this.isEditingLcUsername = false;
    this.lcUsernameControl.setValue(this.currentUser?.lc_username);
  }

  saveLcUsername(): void {
    if (this.lcUsernameControl.invalid || this.isSavingLcUsername) {
      return;
    }

    this.isSavingLcUsername = true;
    this.message = null;

    this.userService.updateUser({ lc_username: this.lcUsernameControl.value })
      .pipe(finalize(() => this.isSavingLcUsername = false))
      .subscribe({
        next: (updatedUser) => {
          if (this.currentUser) {
            this.currentUser.lc_username = updatedUser.lc_username;
          }
          this.isEditingLcUsername = false;
          this.message = { type: 'success', text: 'LeetCode username updated!' };
        },
        error: (err) => {
          console.error("Failed to update LeetCode username:", err);
          this.message = { type: 'error', text: 'Failed to update LeetCode username.' };
        }
      });
  }

  async logout(): Promise<void> {
    await this.auth.logout();
    this.router.navigate(['/login']);
  }
}