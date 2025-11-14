import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';

import { UserService } from 'services/api/user.service';
import { User } from 'models/user.model';
import { Session } from 'models/match';

import { switchMap, map, finalize } from 'rxjs/operators';
import { forkJoin, of } from 'rxjs';

type EnrichedSession = Session & { opponent?: User };

@Component({
  selector: 'app-history-page',
  templateUrl: './history-page.component.html',
  styleUrls: ['./history-page.component.scss'],
  standalone: true,
  imports: [CommonModule, RouterModule]
})
export class HistoryPageComponent implements OnInit {
  userId: number | null = null;
  sessions: EnrichedSession[] = [];
  errorText: string | null = null;
  isLoading = false;

  constructor(
    private userService: UserService
  ) {}

  ngOnInit(): void {
    this.isLoading = true;

    this.userService
      .getMyProfile()
      .pipe(
        switchMap(({ id }) => {
          this.userId = id;
          return this.getSessions(id);
        }),
        switchMap((sessions) => this.addOpponents(sessions)),
        finalize(() => (this.isLoading = false))
      )
      .subscribe({
        next: (data) => (this.sessions = data),
        error: () => (this.errorText = 'Failed to load history')
      });
  }

  private getSessions(userId: number) {
    return this.userService.getUserMatches(userId, 1, 10);
  }

  private addOpponents(sessions: Session[]) {
    if (!sessions.length) return of<EnrichedSession[]>([]);
    const reqs = sessions.map((s) => {
      const oppId = this.findOpponentId(s);
      return oppId
        ? this.userService.getUserProfile(oppId).pipe(map((opponent) => ({ ...s, opponent })))
        : of<EnrichedSession>(s);
    });
    return forkJoin(reqs);
  }

  private findOpponentId(session: Session): number | undefined {
    return session.players.find((pid) => pid !== this.userId!);
  }
}