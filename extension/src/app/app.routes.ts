import { Route } from '@angular/router';
import { LoginPageComponent } from './pages/auth/login-page/login-page.component';
import { LinkLcPageComponent } from './pages/auth/link-lc-page/link-lc-page.component';
import { DashboardPageComponent } from './pages/lobby/dashboard-page/dashboard-page.component';
import { QueuePageComponent } from './pages/lobby/queue-page/queue-page.component';
import { ChallengePageComponent } from './pages/lobby/challenge-page/challenge-page.component';
import { InGamePageComponent } from './pages/game/in-game-page/in-game-page.component';
import { ResultPageComponent } from './pages/game/result-page/result-page.component';
import { SettingsPageComponent } from './pages/settings-page/settings-page.component';
import { AuthGuard } from './core/auth.guard';

export const appRoutes: Route[] = [
  { path: 'login', component: LoginPageComponent },
  { path: '', component: DashboardPageComponent, canActivate: [AuthGuard] },
  { path: 'link-lc',  component: LinkLcPageComponent, canActivate: [AuthGuard] },
  { path: 'settings', component: SettingsPageComponent, canActivate: [AuthGuard] },
  { path: 'queue', component: QueuePageComponent, canActivate: [AuthGuard] },
  { path: 'challenge', component: ChallengePageComponent, canActivate: [AuthGuard] },
  { path: 'game', component: InGamePageComponent, canActivate: [AuthGuard] },
  { path: 'result', component: ResultPageComponent, canActivate: [AuthGuard] },
  { path: '**', redirectTo: '' },
];
