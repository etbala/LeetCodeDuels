import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { lastValueFrom } from 'rxjs';

import { UserInfoResponse } from 'app/models/api_responses';
import { environment } from 'environments/environment';
import { BackgroundService } from 'app/services/background/background.service';

@Component({
  selector: 'app-queue-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './queue-page.component.html',
  styleUrls: ['./queue-page.component.scss']
})
export class QueuePageComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;
  errorText: string | null = null;
  isLoading = true;

  inviteeID: number | null = null;
  opponent: UserInfoResponse | null = null;

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private backgroundService: BackgroundService,
    private http: HttpClient
  ) {}

  ngOnInit(): void {
    const idParam = this.route.snapshot.paramMap.get('inviteeID');
    this.inviteeID = idParam ? +idParam : null;

    if (this.inviteeID == null || Number.isNaN(this.inviteeID)) {
      this.errorText = 'Missing opponent ID.';
      this.isLoading = false;
      return;
    }

    this.loadOpponentProfile(this.inviteeID);
  }

  private async loadOpponentProfile(id: number): Promise<void> {
    this.errorText = null;
    this.isLoading = true;
    try {
      const user = await lastValueFrom(
        this.http.get<UserInfoResponse>(`${this.API_URL}/api/v1/users/${id}`)
      );
      this.opponent = user;
    } catch (err) {
      console.error('Failed to load opponent profile:', err);
      this.errorText = 'Could not load opponent info.';
    } finally {
      this.isLoading = false;
    }
  }

  async cancelInvite(): Promise<void> {
    this.errorText = null;
    try {
      await this.backgroundService.cancelInvitation();
      this.router.navigate(['/']);
    } catch (err) {
      console.error('Failed to cancel invite:', err);
      this.errorText = 'Could not cancel invite. Please try again.';
    }
  }
}
