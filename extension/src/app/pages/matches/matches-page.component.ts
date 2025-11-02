// match-page.component.ts
import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { forkJoin, of, switchMap, map } from 'rxjs';

import { MatchService } from 'services/game-session/game-sessions.service';
import { UserService } from 'services/user/user.service';
import { Session } from 'models/match';
import { User } from 'models/user.model';

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
  players: Record<number, User> = {};
  errorText: string | null = null;

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
      map(m => (this.match = m, m))
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
}
