import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'environments/environment';
import { Difficulty, Problem } from 'app/models/match';
import { Tag } from 'models/tag';

@Injectable({
  providedIn: 'root'
})
export class ProblemsService {
  private readonly apiUrl = `${environment.apiUrl}/api/v1/problems`;

  constructor(private http: HttpClient) {}

  /**
   * Fetches all available problem tags.
   * @returns An Observable of an array of Tag objects.
   * @description Hits the `GET /api/v1/problems/tags` endpoint.
   */
  getTags(): Observable<Tag[]> {
    return this.http.get<Tag[]>(`${this.apiUrl}/tags`);
  }

  /**
   * Fetches a random problem filtered by difficulty and/or tags.
   * @param difficulties - Optional list of difficulties to filter by (e.g., ['Easy', 'Medium']).
   * @param tagIds - Optional list of tag IDs to filter by (e.g., [1, 2, 3]).
   * @returns An Observable of a single Problem object.
   * @description Hits the `GET /api/v1/problems/random?difficulty[]=...&tag[]=...` endpoint.
   */
  getRandomProblem(
    difficulties: Difficulty[] = [],
    tagIds: number[] = []
  ): Observable<Problem> {
    let params = new HttpParams();

    difficulties.forEach(d => {
      params = params.append('difficulty[]', d);
    });

    tagIds.forEach(id => {
      params = params.append('tag[]', id.toString());
    });

    return this.http.get<Problem>(`${this.apiUrl}/random`, { params });
  }
}
