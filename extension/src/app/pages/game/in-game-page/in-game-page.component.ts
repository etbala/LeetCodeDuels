import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { lastValueFrom } from 'rxjs';

import { Session, PlayerSubmission } from 'models/match';
import { UserInfoResponse } from 'models/api_responses';
import { environment } from 'environments/environment';

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

  // player1 / player2 info
  player1?: UserInfoResponse;
  player2?: UserInfoResponse;

  // derived stats
  player1Stats = { submissions: 0, passed: 0, failed: 0 };
  player2Stats = { submissions: 0, passed: 0, failed: 0 };

  isLoading = false;
  errorText: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
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
    // passed / failed from passedTestCases / totalTestCases if present
    const playerSubs = subs.filter(s => s && s.problemID && s);

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
      // get match info
      const match = await lastValueFrom(
        this.http.get<Session>(`${this.API_URL}/api//v1/matches/${this.matchID}`)
      );
      this.matchData = match;

      // assume match.players[0] and match.players[1] exist
      const [p1Id, p2Id] = match.players;

      // load both players' profiles
      const [p1, p2] = await Promise.all([
        this.loadUserProfile(p1Id),
        this.loadUserProfile(p2Id),
      ]);

      this.player1 = p1;
      this.player2 = p2;

      this.player1Stats = this.calcStatsForPlayer(p1Id, match.submissions || []);
      this.player2Stats = this.calcStatsForPlayer(p2Id, match.submissions || []);

    } catch (err) {
      console.error('Failed to load match or player data:', err);
      this.errorText = 'Could not load game data.';
    } finally {
      this.isLoading = false;
    }
  }
}
