import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../environments/environment';
import { Router } from '@angular/router';
import { Tag } from '../../../models/tag';

interface SendInvitationPayload {
  inviteeID: number;
  matchDetails: {
    isRated: boolean;
    difficulties: string[];
    tags: number[];
  };
}

interface Message {
    type: "send_invitation";
    payload: SendInvitationPayload;
}

@Component({
  selector: 'app-dashboard-page',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dashboard-page.component.html',
  styleUrl: './dashboard-page.component.scss'
})
export class DashboardPageComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;
  opponentUsername = '';
  errorMessage: string | null = null;
  difficulties = ['Easy', 'Medium', 'Hard'] as const;
  selectedDifficulties = new Set<string>();

  tags: Tag[] = [];
  selectedTags = new Set<number>();

  constructor(private router: Router, private http: HttpClient) {}

  get selectedDifficultiesList(): string {
    return Array.from(this.selectedDifficulties).join(', ');
  }

  get selectedTagsNames(): string[] {
    return this.tags
      .filter(t => this.selectedTags.has(t.id))
      .map(t => t.name);
  }

  get selectedTagIds(): number[] {
    return Array.from(this.selectedTags);
  }

  ngOnInit() {
    this.http.get<Tag[]>(`${this.API_URL}/api/v1/problems/tags`).subscribe({
      next: data => {
        this.tags = data;
      },
      error: err => console.error('tags fetch error', err)
    });
  }

  toggleDiff(d: string, checked: boolean) {
    if (checked) {
      this.selectedDifficulties.add(d);
    } else {
      this.selectedDifficulties.delete(d);
    }
  }

  toggleTag(id: number, checked: boolean) {
    if (checked) {
      this.selectedTags.add(id);
    } else {
      this.selectedTags.delete(id);
    }
  }

  private async getUserIdFromUsername(input: string): Promise<number | null> {
    const [username, discriminator] = input.split('#');

    try {
      // build query params
      const params = new URLSearchParams();
      params.set('username', username.trim());
      if (discriminator) params.set('discriminator', discriminator.trim());
      params.set('limit', '1'); // only need 1 match

      const userList = await this.http
        .get<{ id: number }[]>(`${this.API_URL}/api/v1/users?${params.toString()}`)
        .toPromise();

      return userList?.[0]?.id ?? null;
    } catch (err) {
      console.error('Error fetching user ID:', err);
      return null;
    }
  }

  private buildMatchPayload(inviteeID: number): Message {
    const payload: SendInvitationPayload = {
      inviteeID,
      matchDetails: {
        isRated: false,
        difficulties: Array.from(this.selectedDifficulties),
        tags: Array.from(this.selectedTags),
      },
    };

    // Return the Message wrapper object
    return {
        type: "send_invitation",
        payload: payload,
    };
  }

  private sendInvitation(message: Message): Promise<void> {
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(`${this.API_URL.replace(/^http/, 'ws')}/ws`);

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        ws.onclose = () => console.log('WebSocket closed');
        reject(new Error('WebSocket connection failed or encountered an error.'));
      };

      ws.onopen = () => {
        console.log('WebSocket connected, sending invitation:', message);
        ws.send(JSON.stringify(message));
        resolve();
      };

      ws.onmessage = (event) => console.log('Message from server:', event.data);
      ws.onclose = () => console.log('WebSocket closed');
    });
  }

  async startDuel() {
    this.errorMessage = null; // reset before each attempt

    const trimmedUsername = this.opponentUsername.trim();
    if (!trimmedUsername) {
      this.errorMessage = 'Please enter an opponent username.';
      return;
    }

    // TODO: Remove this when implmenting functonality
    this.router.navigate(['/queue', trimmedUsername]);

    try {
      const inviteeID = await this.getUserIdFromUsername(trimmedUsername);
      if (!inviteeID) {
        this.errorMessage = 'Opponent not found. Make sure username#id is correct.';
        console.error('Could not find user ID for opponent');
        return;
      }

      const payload = this.buildMatchPayload(inviteeID);
      try {
        await this.sendInvitation(payload);
        this.router.navigate(['/queue', trimmedUsername]);
      } catch (inviteErr) {
        this.errorMessage = 'Failed to send duel invitation. Please try again.';
        console.error('Error sending invitation:', inviteErr);
      }

    } catch (err) {
      this.errorMessage = 'An unexpected error occurred. Please try again.';
      console.error('Error starting duel:', err);
    }
  }
}
