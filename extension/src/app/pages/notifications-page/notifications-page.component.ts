import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { NotificationsResponse, InviteNotification } from 'app/models/api_responses';
// import { BackgroundAction, BackgroundActionType, AcceptInvitationPayload, DeclineInvitationPayload } from '../../models/background-actions';

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

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.invites$ = this.fetchPendingInvites();
  }

  private fetchPendingInvites(): Observable<InviteNotification[]> {
    const fullUrl = `${this.API_URL}/api/v1/me/notifications`;

    return this.http.get<NotificationsResponse>(fullUrl).pipe(
      map(response => response.invites || [])
    );
  }

  public acceptInvite(invite: InviteNotification): void {
    console.log(`Accepted invite from: ${invite.from_user.username}`);
  }

  public declineInvite(invite: InviteNotification): void {
    console.log(`Declined invite from: ${invite.from_user.username}`);
  }
}