import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-queue-page',
  templateUrl: './queue-page.component.html',
  styleUrl: './queue-page.component.scss'
})
export class QueuePageComponent implements OnInit {
  opponentUsername: string | null = null;

  constructor(private router: Router, private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.opponentUsername = this.route.snapshot.paramMap.get('username');
  }

  async cancelInvite() {
    this.router.navigate(['/']);
  }
}