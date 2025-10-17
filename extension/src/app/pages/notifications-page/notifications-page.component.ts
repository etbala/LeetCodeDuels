import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, of } from 'rxjs'; // TODO: Delete of with mock data deletion
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { NotificationsResponse, InviteNotification } from 'app/models/api_responses';
import { Difficulty } from 'app/models/match'; // TODO: Delete with mock data deletion
// import { BackgroundAction, BackgroundActionType, AcceptInvitationPayload, DeclineInvitationPayload } from '../../models/background-actions';

// TODO: Delete when implment functionality. Used for match page navigation
const MOCK_INVITES: InviteNotification[] = [
  {
    from_user: {
      id: 101,
      username: 'CodeSlayer_42',
      discriminator: '001',
      lc_username: 'codeslayer',
      avatar_url: 'https://placehold.co/100x100/7B68EE/FFFFFF?text=CS',
      rating: 1850,
    },
    matchDetails: {
      isRated: true,
      difficulties: [Difficulty.Medium, Difficulty.Hard],
      tags: [1, 12, 25],
    },
    createdAt: new Date().toISOString(),
  },
  {
    from_user: {
      id: 202,
      username: 'AlgoQueen',
      discriminator: '002',
      lc_username: 'algoqueen',
      avatar_url: 'https://placehold.co/100x100/F76D82/FFFFFF?text=AQ',
      rating: 2100,
    },
    matchDetails: {
      isRated: false,
      difficulties: [Difficulty.Easy],
      tags: [5, 18],
    },
    createdAt: new Date(Date.now() - 3600 * 1000).toISOString(), 
  },
];

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

  constructor(private router: Router, private http: HttpClient) {}

  ngOnInit(): void {
    // this.invites$ = this.fetchPendingInvites();
    this.invites$ = of(MOCK_INVITES);
  }

  private fetchPendingInvites(): Observable<InviteNotification[]> {
    const fullUrl = `${this.API_URL}/api/v1/me/notifications`;

    return this.http.get<NotificationsResponse>(fullUrl).pipe(
      map(response => response.invites || [])
    );
  }

  public acceptInvite(invite: InviteNotification): void {
    console.log(`Navigating to game, accepting invite from ${invite.from_user.username}`);
    this.router.navigate(['/game']);
  }

  public declineInvite(invite: InviteNotification): void {
    console.log(`Declined invite from ${invite.from_user.username}`);
  }
}