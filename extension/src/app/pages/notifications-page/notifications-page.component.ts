import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

// TODO: Move these interfaces to a separate `notifications.model.ts` file later.
// --- Start of Interfaces ---

export interface UserInfo {
  id: string;
  username: string;
  // Add any other user fields you expect from the API
}

export interface MatchDetails {
  isRated: boolean;
  difficulties: string[]; // e.g., ["Easy", "Medium"]
  tags: number[];
}

export interface InviteNotification {
  from_user: UserInfo;
  matchDetails: MatchDetails;
  createdAt: string; // ISO 8601 date string
}

export interface NotificationsResponse {
  invites: InviteNotification[];
}

// --- End of Interfaces ---


@Component({
  selector: 'app-notification-page',
  templateUrl: './notifications-page.component.html',
  styleUrls: ['./notifications-page.component.scss']
})
export class NotificationPageComponent implements OnInit {
  
  // Using an observable with the '$' suffix is a common convention in Angular
  // It will hold the stream of invites coming from the service.
  public invites$!: Observable<InviteNotification[]>;

  // --- TODO: This API logic should be moved to a dedicated service later. ---
  private apiUrl = '/api/users/me/notifications'; 

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    // When the component initializes, we call the local method to fetch the invites.
    this.invites$ = this.fetchPendingInvites();
  }

  private fetchPendingInvites(): Observable<InviteNotification[]> {
    return this.http.get<NotificationsResponse>(this.apiUrl).pipe(
      map(response => response.invites || []) // Use empty array as a fallback
    );
  }
  // --- End of API logic ---

  // Placeholder for accepting an invite
  acceptInvite(invite: InviteNotification): void {
    console.log('Accepting invite from:', invite.from_user.username);
    // Future logic: Call a service method to accept the invite via API
  }

  // Placeholder for declining an invite
  declineInvite(invite: InviteNotification): void {
    console.log('Declining invite from:', invite.from_user.username);
    // Future logic: Call a service method to decline the invite via API
  }
}