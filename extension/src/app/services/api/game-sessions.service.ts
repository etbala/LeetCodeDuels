import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Session, PlayerSubmission } from 'models/match';

@Injectable({
    providedIn: 'root'
})
export class MatchService {
  private readonly apiUrl = environment.apiUrl + '/api/v1/matches';

  constructor(private http: HttpClient) {}

  /**
   * Fetches detailed information about a specific match.
   * @param id - The unique ID of the match.
   * @returns An Observable containing a `Session` object with full match details.
   * @description Calls the `GET /api/v1/matches/{id}` endpoint.
   */
  getMatch(id: string | number): Observable<Session> {
    return this.http.get<Session>(`${this.apiUrl}/${id}`);
  }

  /**
   * Fetches all submissions associated with a specific match.
   * @param id - The unique ID of the match.
   * @returns An Observable array of PlayerSubmission objects.
   * @description Calls `GET /api/v1/matches/{id}/submissions` to retrieve
   *              each player's submissions (pass/fail counts, runtime, etc.).
   */
  getMatchSubmissions(id: string | number): Observable<PlayerSubmission[]> {
    return this.http.get<PlayerSubmission[]>(`${this.apiUrl}/${id}/submissions`);
  }
}
