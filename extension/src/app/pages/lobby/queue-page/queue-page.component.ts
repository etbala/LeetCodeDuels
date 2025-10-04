import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-queue-page',
  imports: [],
  templateUrl: './queue-page.component.html',
  styleUrl: './queue-page.component.scss'
})
export class QueuePageComponent {
  constructor(private router: Router) {}

  async cancelInvite() {
    this.router.navigate(['/']);
  }
}
