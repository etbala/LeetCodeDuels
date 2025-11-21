import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { lastValueFrom, Subject, merge, Subscription, timer } from 'rxjs';
import { map, takeUntil, tap, switchMap } from 'rxjs/operators';

import { UserInfoResponse } from 'app/models/api_responses';
import { environment } from 'environments/environment';
import { BackgroundService } from 'app/services/background/background.service';
import { ExtensionEventType } from 'models/extension-events';
import { ExtensionEventsService } from 'services/background/extension-events.service';

@Component({
  selector: 'app-queue-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './queue-page.component.html',
  styleUrls: ['./queue-page.component.scss']
})
export class QueuePageComponent implements OnInit, OnDestroy {
  private readonly API_URL = environment.apiUrl;
  private destroy$ = new Subject<void>();
  private queueTimeoutSub?: Subscription;
  errorText: string | null = null;
  isLoading = true;

  inviteeID: number | null = null;
  opponent: UserInfoResponse | null = null;

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private backgroundService: BackgroundService,
    private http: HttpClient,
    private events: ExtensionEventsService
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
    this.handleStartGameEvents();
    this.handleInviteTerminalEvents();
    this.startQueueTimeout(); // 2-minute safety fallback
  }

  // Navigate to /game when StartGame arrives (and matches invitee if present).
  private handleStartGameEvents(): void {
    this.events
      .listenFor<{ opponentID?: number; matchId?: string }>(ExtensionEventType.StartGame)
      .pipe(takeUntil(this.destroy$))
      .subscribe(({ opponentID, matchId }) => {
        if (!this.inviteeID || opponentID === this.inviteeID) {
          this.clearQueueTimeout();
          this.router.navigate(matchId ? ['/game', matchId] : ['/game']);
        }
      });
  }

  // Show message for invitation decline / fail, then return to dashboard.
  private handleInviteTerminalEvents(): void {
    merge(
      this.events
        .listenFor<void>(ExtensionEventType.InvitationDeclined)
        .pipe(map(() => 'Your invitation was declined.')),
      this.events
        .listenFor<void>(ExtensionEventType.InvitationCanceled)
        .pipe(map(() => 'Your invitation was canceled.')),
      this.events
        .listenFor<void>(ExtensionEventType.UserOffline)
        .pipe(map(() => 'User went offline.')),
      this.events
        .listenFor<void>(ExtensionEventType.InvitationNonexistant)
        .pipe(map(() => 'Invitation no longer exists.'))
    )
      .pipe(
        takeUntil(this.destroy$),
        tap(msg => {
          this.clearQueueTimeout();
          this.errorText = msg;
          this.isLoading = false;
        }),
        switchMap(() => timer(1200)) // brief notice, then bounce
      )
      .subscribe(() => this.router.navigate(['/']));
  }

  // Start a 2-minute safety timeout to avoid waiting forever.
  private startQueueTimeout(): void {
    this.clearQueueTimeout();
    this.queueTimeoutSub = timer(120000) // 2 minutes
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        this.errorText = 'No response — taking you back to the dashboard.';
        this.isLoading = false;
        this.router.navigate(['/']);
      });
  }

  private clearQueueTimeout(): void {
    if (this.queueTimeoutSub) {
      this.queueTimeoutSub.unsubscribe();
      this.queueTimeoutSub = undefined;
    }
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

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
