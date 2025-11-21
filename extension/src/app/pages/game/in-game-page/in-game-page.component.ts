import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { lastValueFrom, Subject, takeUntil } from 'rxjs';

import { Session, PlayerSubmission } from 'models/match';
import { UserInfoResponse } from 'models/api_responses';
import { environment } from 'environments/environment';
import { BackgroundService } from 'services/background/background.service';
import { MatchService } from 'services/api/game-sessions.service';
import { ExtensionEventsService } from 'services/background/extension-events.service';
import { ExtensionEventType } from 'models/extension-events';

@Component({
  selector: 'app-in-game-page',
  imports: [CommonModule],
  templateUrl: './in-game-page.component.html',
  styleUrls: ['./in-game-page.component.scss'],
})
export class InGamePageComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;
  private destroy$ = new Subject<void>();

  matchID!: string;
  matchData?: Session;
  problemTitle = '';
  problemLink: string | null = null;

  // player1 / player2 info
  player1?: UserInfoResponse;
  player2?: UserInfoResponse;

  // derived stats
  player1Stats = { submissions: 0, passed: 0, failed: 0 };
  player2Stats = { submissions: 0, passed: 0, failed: 0 };

  isConfirmingForfeit = false;
  isLoading = false;
  errorText: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private http: HttpClient,
    private backgroundService: BackgroundService,
    private matchService: MatchService,
    private extensionEvents: ExtensionEventsService,
  ) {}

  async ngOnInit(): Promise<void> {
    this.matchID = this.route.snapshot.paramMap.get('matchID')!;
    this.setupEventListeners();
    await this.loadMatchAndPlayers();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private setupEventListeners(): void {
    // Any submission event from server -> refresh stats / match data
    this.extensionEvents
      .listenFor<any>(ExtensionEventType.OpponentSubmission)
      .pipe(takeUntil(this.destroy$))
      .subscribe(payload => {
        // only react if it's for THIS match
        const sessionId = (payload as any)?.sessionID ?? (payload as any)?.matchID;
        if (sessionId && String(sessionId) !== String(this.matchID)) {
          return;
        }

        this.loadMatchAndPlayers();
      });

    // Game over so go to match-over page
    this.extensionEvents
      .listenFor<any>(ExtensionEventType.GameOver)
      .pipe(takeUntil(this.destroy$))
      .subscribe(async payload => {
        const sessionId = (payload as any)?.sessionID ?? this.matchID;
        try {
          await chrome.storage.local.remove('lastSession');
        } catch (e) {
          console.warn('[InGame] Failed to clear lastSession from storage:', e);
        }
        this.router.navigate(['/match-over', sessionId]);
      });
  }

  private async loadUserProfile(id: number): Promise<UserInfoResponse> {
    return await lastValueFrom(
      this.http.get<UserInfoResponse>(`${this.API_URL}/api/v1/users/${id}`)
    );
  }

  private calcStatsForPlayer(playerId: number, subs: PlayerSubmission[]) {
    if (!Array.isArray(subs)) {
      return { submissions: 0, passed: 0, failed: 0 };
    }

    const playerSubs = subs.filter(s => s.playerID === playerId);
    const submissions = playerSubs.length;

    if (submissions === 0) {
      return { submissions: 0, passed: 0, failed: 0 };
    }

    // sort by time and take the latest with test case info
    const withCases = playerSubs
      .filter(s => s.passedTestCases !== undefined && s.totalTestCases !== undefined)
      .sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime());

    if (withCases.length === 0) {
      return { submissions, passed: 0, failed: 0 };
    }

    const latest = withCases[withCases.length - 1];
    const passed = latest.passedTestCases!;
    const failed = latest.totalTestCases! - latest.passedTestCases!;

    return { submissions, passed, failed };
  }

  private async loadMatchAndPlayers(): Promise<void> {
    this.isLoading = true;
    this.errorText = null;

    try {
      const [match, submissions] = await Promise.all([
        lastValueFrom(this.matchService.getMatch(this.matchID)),
        lastValueFrom(this.matchService.getMatchSubmissions(this.matchID)),
      ]);

      this.matchData = match;
      const [p1Id, p2Id] = match.players;

      const [p1, p2] = await Promise.all([
        this.loadUserProfile(p1Id),
        this.loadUserProfile(p2Id),
      ]);

      this.player1 = p1;
      this.player2 = p2;

      this.player1Stats = this.calcStatsForPlayer(p1Id, submissions);
      this.player2Stats = this.calcStatsForPlayer(p2Id, submissions);

      this.problemTitle = match.problem?.name || match.problem?.slug || 'Unknown Problem';
      this.problemLink = match.problem?.slug
        ? `https://leetcode.com/problems/${match.problem.slug}/`
        : null;
    } catch (err) {
      console.error('Failed to load match or player data:', err);
      this.errorText = 'Could not load game data.';
    } finally {
      this.isLoading = false;
    }
  }

  async forfeitDuel(): Promise<void> {
    this.errorText = null;

    // grab the session id before forfeiting
    const { lastSession } = await chrome.storage.local.get('lastSession') as { lastSession?: Session };
    const sessionId = lastSession?.sessionID;

    try {
      await this.backgroundService.forfeitDuel();

      // prevent future auto-redirects
      await chrome.storage.local.remove('lastSession');

      // go straight to match-over
      if (sessionId) {
        await this.router.navigate(['/match-over', sessionId]);
      } else {
        await this.router.navigate(['/']);
      }
    } catch (err) {
      console.error('Failed to forfeitDuel:', err);
      this.errorText = 'Could not forfeit duel. Please try again.';
    }
  }
}
