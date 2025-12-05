import { Component, OnInit, ViewChild, ElementRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

import { BackgroundService } from 'app/services/background/background.service';
import { SendInvitationPayload } from 'app/models/background-actions';
import { Difficulty, Session } from 'app/models/match';

import { Tag } from 'models/tag';
import { UserService } from 'services/api/user.service';
import { MatchService } from 'services/api/game-sessions.service';
import { ProblemsService } from 'services/api/problems.service';

import { firstValueFrom } from 'rxjs';

const isSessionCompleted = (s?: Session | null) =>
  !!s && (
    s.status === 'Won' ||
    s.status === 'Canceled' ||
    s.status === 'Reverted' ||
    s.winner > 0 ||
    !!s.endTime
  );

@Component({
  selector: 'app-dashboard-page',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dashboard-page.component.html',
  styleUrl: './dashboard-page.component.scss'
})
export class DashboardPageComponent implements OnInit {
  opponentUsername = '';
  currentUserId: number | null = null;
  errorMessage: string | null = null;
  difficulties = ['Easy', 'Medium', 'Hard'] as const;
  selectedDifficulties = new Set<string>();
  isLoading = true;

  tags: Tag[] = [];
  selectedTags = new Set<number>();
  tagFilter = '';
  @ViewChild('filterInput') filterInput!: ElementRef<HTMLInputElement>;

  constructor(
    private router: Router,
    private backgroundService: BackgroundService,
    private userService: UserService,
    private matchService: MatchService,
    private problemsService: ProblemsService,
  ) {}

  get selectedDifficultiesList(): string {
    return Array.from(this.selectedDifficulties).join(', ');
  }

  get selectedTagsNames(): string[] {
    return this.tags
      .filter(t => this.selectedTags.has(t.id))
      .map(t => t.name);
  }

  get selectedTagIds(): number[] {
    return Array.from(this.selectedTags);
  }

  get filteredTags() {
    const q = this.tagFilter.trim().toLowerCase();
    if (!q) return this.tags;
    return this.tags.filter(t => t.name.toLowerCase().includes(q));
  }

  ngOnInit() {
    this.getUserIdAndRedirect();

    this.problemsService.getTags().subscribe({
      next: tags => {
        this.tags = tags;
      },
      error: err => console.error('tags fetch error', err)
    });
  }

  getUserIdAndRedirect() {
    this.userService.getMyProfile().subscribe({
      next: user => {
        this.currentUserId = user.id;

        if (!user.lc_username) {
          this.errorMessage =
            'Heads up: your LeetCode username is not configured. Submissions may not be tracked.';
        }

        this.redirectIfInGame();
      },
      error: err => {
        console.error('profile fetch error', err);
        this.isLoading = false;
      }
    });
  }

  async checkJustFinishedMatch() {
    try {
      const { lastSession } = await chrome.storage.local.get('lastSession') as {
        lastSession?: Session;
      };

      if (!lastSession) return;

      const current = await firstValueFrom(
        this.matchService.getMatch(lastSession.sessionID)
      );

      if (isSessionCompleted(current)) {
        await chrome.storage.local.remove('lastSession');
        this.router.navigate(['/match-over', current.sessionID]);
      }
    } catch (err) {
      console.error('checkJustFinishedMatch error', err);
    }
  }

  redirectIfInGame() {
    if (!this.currentUserId) {
      this.isLoading = false;
      return;
    }

    this.userService.getUserStatus(this.currentUserId).subscribe({
      next: async res => {
        if (res.in_game && res.game_id) {
          const current = await firstValueFrom(
            this.matchService.getMatch(res.game_id)
          );
          await chrome.storage.local.set({ lastSession: current });

          this.router.navigate(['/game', res.game_id]);
        } else {
          await this.checkJustFinishedMatch();
        }

        this.isLoading = false;
      },
      error: err => {
        console.error('status check error', err);
        this.isLoading = false;
      }
    });
  }

  toggleDiff(d: string, checked: boolean) {
    if (checked) {
      this.selectedDifficulties.add(d);
    } else {
      this.selectedDifficulties.delete(d);
    }
  }

  toggleTag(id: number, checked: boolean) {
    if (checked) {
      this.selectedTags.add(id);
    } else {
      this.selectedTags.delete(id);
    }
  }

  onDropdownToggle(event: Event) {
    const details = event.target as HTMLDetailsElement;
    if (details.open) {
      // wait a tick to let the dropdown render before focusing
      setTimeout(() => this.filterInput?.nativeElement.focus(), 50);
    }
  }

  private getUserIdFromUsername(input: string): Promise<number | null> {
    return firstValueFrom(this.userService.findUserId(input));
  }

  private buildMatchPayload(inviteeID: number): SendInvitationPayload {
    return {
      inviteeID,
      matchDetails: {
        isRated: false,
        difficulties: Array.from(this.selectedDifficulties).map(s => s as Difficulty),
        tags: Array.from(this.selectedTags),
      },
    };
  }

  private sendInvitation(payload: SendInvitationPayload) {
    return this.backgroundService.sendInvitation(payload);
  }

  async startDuel() {
    this.errorMessage = null;

    const trimmedUsername = this.opponentUsername.trim();
    if (!trimmedUsername) {
      this.errorMessage = 'Please enter an opponent username.';
      return;
    }

    try {
      const inviteeID = await this.getUserIdFromUsername(trimmedUsername);
      if (!inviteeID) {
        this.errorMessage = 'Opponent not found. Make sure username#id is correct.';
        console.error('Could not find user ID for opponent');
        return;
      }

      const payload = this.buildMatchPayload(inviteeID);
      try {
        await this.sendInvitation(payload);
        this.router.navigate(['/queue', inviteeID]);
      } catch (inviteErr) {
        this.errorMessage = 'Failed to send duel invitation. Please try again.';
        console.error('Error sending invitation:', inviteErr);
      }
    } catch (err) {
      this.errorMessage = 'An unexpected error occurred. Please try again.';
      console.error('Error starting duel:', err);
    }
  }
}
