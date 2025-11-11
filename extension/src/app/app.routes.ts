import { Route } from '@angular/router';
import { LoginPageComponent } from './pages/auth/login-page/login-page.component';
import { DashboardPageComponent } from './pages/lobby/dashboard-page/dashboard-page.component';
import { QueuePageComponent } from './pages/lobby/queue-page/queue-page.component';
import { ChallengePageComponent } from './pages/lobby/challenge-page/challenge-page.component';
import { InGamePageComponent } from './pages/game/in-game-page/in-game-page.component';
import { ResultPageComponent } from './pages/game/result-page/result-page.component';
import { SettingsPageComponent } from './pages/settings-page/settings-page.component';
import { NotificationPageComponent } from './pages/notifications-page/notifications-page.component';
import { HistoryPageComponent } from 'pages/history/history-page.component';
import { MatchesPageComponent } from 'pages/matches/matches-page.component';
import { MatchOverPageComponet } from './pages/match-over-page/match-over-page.component';
import { AuthGuard } from './services/auth/auth.guard';

export const appRoutes: Route[] = [
  { path: 'login', component: LoginPageComponent },
  { path: '', component: DashboardPageComponent },
  { path: 'settings', component: SettingsPageComponent, canActivate: [AuthGuard] },
  { path: 'notifications', component: NotificationPageComponent, canActivate: [AuthGuard] },
  { path: 'history', component: HistoryPageComponent, canActivate: [AuthGuard] },
  { path: 'match-over/:id', component: MatchOverPageComponet, canActivate: [AuthGuard] },
  { path: 'matches/:id', component: MatchesPageComponent, canActivate: [AuthGuard] },
  { path: 'queue/:inviteeID', component: QueuePageComponent, canActivate: [AuthGuard] },
  { path: 'challenge', component: ChallengePageComponent, canActivate: [AuthGuard] },
  { path: 'game/:matchID', component: InGamePageComponent, canActivate: [AuthGuard] },
  { path: 'result', component: ResultPageComponent, canActivate: [AuthGuard] },
  { path: '**', redirectTo: '' },
];
