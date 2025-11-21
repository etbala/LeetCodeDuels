// Angular Imports
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

// RxJS Imports
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

// Application Imports
import { BackgroundService } from 'app/services/background/background.service';
import { NotificationsResponse, InviteNotification } from 'app/models/api_responses';
import { AcceptInvitationPayload, DeclineInvitationPayload } from '../../models/background-actions';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-notification-page',
  templateUrl: './notifications-page.component.html',
  styleUrls: ['./notifications-page.component.scss'],
  standalone: true,
  imports: [CommonModule]
})
export class NotificationPageComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;
  public invites$!: Observable<InviteNotification[]>;
  errorText: string | null = null;

  constructor(
    private router: Router,
    private http: HttpClient,
    private backgroundService: BackgroundService
  ) {}

  ngOnInit(): void {
    this.invites$ = this.fetchPendingInvites();
  }

  private fetchPendingInvites(): Observable<InviteNotification[]> {
    const fullUrl = `${this.API_URL}/api/v1/users/me/notifications`;

    return this.http.get<NotificationsResponse>(fullUrl).pipe(
      map(response => response.invites || [])
    );
  }

  public async acceptInvite(invite: InviteNotification): Promise<void> {
    this.errorText = null;
    const payload: AcceptInvitationPayload = { inviterID: invite.from_user.id };

    try {
      await this.backgroundService.acceptInvitation(payload);
      console.log(`Navigating to game, accepting invite from ${invite.from_user.username}`);
      this.router.navigate(['/game']);
    } catch (err) {
      console.error('Failed to accept invite:', err);
      this.errorText = 'Could not accept invite. Please try again.';
    }
  }

  public async declineInvite(invite: InviteNotification): Promise<void> {
    this.errorText = null;
    const payload: DeclineInvitationPayload = { inviterID: invite.from_user.id };

    try {
      await this.backgroundService.declineInvitation(payload);
      this.invites$ = this.fetchPendingInvites();
    } catch (err) {
      console.error('Failed to decline invite:', err);
      this.errorText = 'Could not decline invite. Please try again.';
    }
  }
}