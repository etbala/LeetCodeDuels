import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { lastValueFrom } from 'rxjs';

import { Session, PlayerSubmission } from 'models/match';
import { UserInfoResponse } from 'models/api_responses';
import { environment } from 'environments/environment';
import { BackgroundService } from 'services/background/background.service';
import { MatchService } from 'services/api/game-sessions.service';

@Component({
  selector: 'app-in-game-page',
  imports: [CommonModule],
  templateUrl: './in-game-page.component.html',
  styleUrls: ['./in-game-page.component.scss'],
})
export class InGamePageComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;

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
  ) {}

  async ngOnInit(): Promise<void> {
    this.matchID = this.route.snapshot.paramMap.get('matchID')!;
    await this.loadMatchAndPlayers();
  }

  private async loadUserProfile(id: number): Promise<UserInfoResponse> {
    return await lastValueFrom(
      this.http.get<UserInfoResponse>(`${this.API_URL}/api/v1/users/${id}`)
    );
  }

  private calcStatsForPlayer(playerId: number, subs: PlayerSubmission[]) {
    const playerSubs = subs.filter(s => s.playerID === playerId);

    const submissions = playerSubs.length;
    let passed = 0;
    let failed = 0;

    for (const s of playerSubs) {
      if (s.passedTestCases !== undefined && s.totalTestCases !== undefined) {
        passed += s.passedTestCases;
        failed += (s.totalTestCases - s.passedTestCases);
      }
    }

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
