import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../environments/environment.prod';

interface Tag {
  id: number;
  name: string;
}

@Component({
  selector: 'app-dashboard-page',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dashboard-page.component.html',
  styleUrl: './dashboard-page.component.scss'
})
export class DashboardPageComponent implements OnInit {
  opponentUsername = '';
  difficulties = ['Easy', 'Medium', 'Hard'] as const;
  selectedDifficulties = new Set<string>();
  private readonly API_URL = environment.apiUrl;

  tags: Tag[] = [];
  selectedTags = new Set<number>();

  constructor(private http: HttpClient) {}

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
        console.log('tags response', data);
        this.tags = data;
      },
      error: err => console.error('tags fetch error', err)
    });
  }

  toggleDiff(d: string, checked: boolean) {
    checked ? this.selectedDifficulties.add(d) : this.selectedDifficulties.delete(d);
  }

  toggleTag(id: number, checked: boolean) {
    checked ? this.selectedTags.add(id) : this.selectedTags.delete(id);
  }

  startDuel() {
    const payload = {
      opponent: this.opponentUsername.trim(),
      difficulties: Array.from(this.selectedDifficulties),
      tags: Array.from(this.selectedTags),
      isRated: false,
    };
    console.log('StartDuel payload:', payload);
  }
}