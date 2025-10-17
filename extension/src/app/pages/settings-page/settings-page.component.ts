import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { finalize } from 'rxjs/operators';
import { AuthService } from '../../services/auth/auth.service';
import { UserService } from '../../services/user/user.service';
import { User } from 'app/models/user.model';

@Component({
  selector: 'app-settings-page',
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './settings-page.component.html',
  styleUrl: './settings-page.component.scss'
})
export class SettingsPageComponent implements OnInit {
  profileForm: FormGroup;
  currentUser: User | null = null;
  
  isSaving = false;
  isLoading = true;
  message: { type: 'success' | 'error', text: string } | null = null;
  
  constructor(
    private auth: AuthService,
    private router: Router,
    private fb: FormBuilder,
    private userService: UserService
  ) {
    this.profileForm = this.fb.group({
      username: ['', [Validators.required, Validators.maxLength(32)]],
      lc_username: ['', [Validators.maxLength(50)]]
    });
  }

  ngOnInit(): void {
    // Fetch the current user's data from the API
    this.userService.getMyProfile()
      .pipe(finalize(() => this.isLoading = false))
      .subscribe({
        next: (user) => {
          this.currentUser = user;
          this.profileForm.patchValue({
            username: user.username,
            lc_username: user.lc_username
          });
        },
        error: (err) => {
          console.error("Failed to load user profile:", err);
          this.message = { type: 'error', text: 'Could not load your profile. Please refresh.' };
        }
      });
  }

  saveProfile(): void {
    if (this.profileForm.invalid || this.isSaving) {
      return;
    }
    
    this.isSaving = true;
    this.message = null;
    const updatedData = this.profileForm.value;
    
    // Call the updateUser method from the UserService
    this.userService.updateUser(updatedData)
      .pipe(finalize(() => this.isSaving = false))
      .subscribe({
        next: () => {
          this.message = { type: 'success', text: 'Profile updated successfully!' };
          this.profileForm.markAsPristine(); // Resets the form's 'dirty' state
        },
        error: (err) => {
          console.error("Failed to update profile:", err);
          // You could inspect `err.error.message` for a more specific message from the backend
          this.message = { type: 'error', text: 'Failed to update profile. The display name may be taken.' };
        }
      });
  }

  deleteAccount(): void {
    const confirmation = window.confirm('Are you sure you want to delete your account? This action cannot be undone.');
    if (confirmation) {
      this.message = null;
      this.userService.deleteUser().subscribe({
        next: async () => {
          await this.logout();
        },
        error: (err) => {
          console.error("Failed to delete account:", err);
          this.message = { type: 'error', text: 'Failed to delete account. Please try again.' };
        }
      });
    }
  }

  async logout(): Promise<void> {
    await this.auth.logout();
    this.router.navigate(['/login']);
  }
}