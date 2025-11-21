// match-page.component.ts
import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { forkJoin, of, switchMap, map } from 'rxjs';

import { MatchService } from 'services/api/game-sessions.service';
import { UserService } from 'services/api/user.service';
import { Session } from 'models/match';
import { User } from 'models/user.model';
import { PlayerSubmission } from 'models/match';

@Component({
  selector: 'app-match-page',
  templateUrl: './matches-page.component.html',
  styleUrls: ['./matches-page.component.scss'],
  standalone: true,
  imports: [CommonModule]
})
export class MatchesPageComponent implements OnInit {
  userId: number | null = null;
  match?: Session;
  players: Record<number, User | undefined> = {};
  errorText: string | null = null;
  statsByPlayerId: Record<number, {
    submissions: number;
    passed: number;
    failed: number
  }> = {};

  constructor(
    private route: ActivatedRoute,
    private matchService: MatchService,
    private userService: UserService
  ) {}

  ngOnInit(): void {
    this.loadMe()
      .pipe(
        switchMap(() => this.routeId()),
        switchMap((id) => this.getMatch(id)),
        switchMap((m) => this.loadPlayers(m))
      )
      .subscribe({
        next: (entries) => this.setPlayers(entries),
        error: () => (this.errorText = 'Failed to load match')
      });
  }

  private loadMe() {
    return this.userService.getMyProfile().pipe(
      map(({ id }) => (this.userId = id, id))
    );
  }

  private routeId() {
    return this.route.paramMap.pipe(map(p => p.get('id')!));
  }

  private getMatch(id: string) {
    return this.matchService.getMatch(id).pipe(
      map(m => {
        this.match = m;
        this.computeStats(m);
        return m;
      })
    );
  }

  private loadPlayers(match: Session) {
    const reqs = match.players.map(pid =>
      this.userService.getUserProfile(pid).pipe(map(u => [pid, u] as const))
    );
    return reqs.length ? forkJoin(reqs) : of([]);
  }

  private setPlayers(entries: readonly (readonly [number, User])[]) {
    entries.forEach(([id, u]) => (this.players[id] = u));
  }

  private computeStats(match: Session) {
    const subs = match.submissions || [];
    this.statsByPlayerId = {};

    for (const pid of match.players) {
      this.statsByPlayerId[pid] = this.calcStatsForPlayer(pid, subs);
    }
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
}
