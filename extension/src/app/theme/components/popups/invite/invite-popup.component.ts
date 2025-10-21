import { Component, OnInit, OnDestroy } from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { ExtensionEventsService } from 'services/background/extension-events.service';
import { BackgroundService } from 'services/background/background.service';
import { InvitationRequestPayload, ExtensionEventType } from 'models/extension-events';
import { AcceptInvitationPayload, DeclineInvitationPayload } from 'models/background-actions';
import { CommonModule } from '@angular/common';
import { UserService } from 'app/services/user/user.service';

@Component({
  selector: 'app-invite-popup',
  imports: [CommonModule],
  templateUrl: './invite-popup.component.html',
  styleUrls: ['./invite-popup.component.scss'],
})
export class InvitePopupComponent implements OnInit, OnDestroy {

  private destroy$ = new Subject<void>();
  public receivedInvitation: InvitationRequestPayload | null = null;

  constructor(
    private extensionEvents: ExtensionEventsService,
    private backgroundService: BackgroundService,
    private userService: UserService
  ) {}

  ngOnInit(): void {
    this.extensionEvents
      .listenFor<InvitationRequestPayload>(ExtensionEventType.InvitationRequest)
      .pipe(takeUntil(this.destroy$))
      .subscribe((payload) => {
        console.log('Invitation received in component:', payload);
        this.receivedInvitation = payload;
        // todo: get user info for inviterID and display username/discriminator/avatar
      });

    this.extensionEvents
      .listenFor<void>(ExtensionEventType.InvitationCanceled)
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
          console.log('Invitation was canceled.');
          this.receivedInvitation = null;
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  // Accepts the current invitation and sends message to service worker.
  async accept(): Promise<void> {
    if (!this.receivedInvitation) return;

    try {
      const payload: AcceptInvitationPayload = { 
        inviterID: this.receivedInvitation.inviterID
      };
      await this.backgroundService.acceptInvitation(payload);
      console.log('Accept invitation message sent.');
    } catch (error) {
      console.error('Failed to accept invitation:', error);
    } finally {
      this.receivedInvitation = null;
    }
  }

  // Declines the current invitation and sends message to service worker.
  async decline(): Promise<void> {
    if (!this.receivedInvitation) return;

    try {
      const payload: DeclineInvitationPayload = { 
        inviterID: this.receivedInvitation.inviterID
      };
      await this.backgroundService.declineInvitation(payload);
      console.log('Decline invitation message sent.');
    } catch (error) {
      console.error('Failed to decline invitation:', error);
    } finally {
      this.receivedInvitation = null;
    }
  }
}