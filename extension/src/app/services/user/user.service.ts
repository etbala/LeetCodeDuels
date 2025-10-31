import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'environments/environment';
import { User } from 'app/models/user.model';
import { Session } from 'models/match';

export interface UpdateUserRequest {
  username?: string;
  lc_username?: string;
}

export interface UpdateUserResponse {
  id: number;
  username: string;
  discriminator: string;
  lc_username: string;
}

@Injectable({
  providedIn: 'root'
})
export class UserService {
  private readonly apiUrl = environment.apiUrl + '/api/v1/users';

  constructor(private http: HttpClient) {}

  /**
   * Fetches the profile information for the currently authenticated user.
   * @returns An Observable of the user's profile data.
   * @description Hits the `GET /api/v1/users/me` endpoint.
   */
  getMyProfile(): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/me`);
  }

  /**
   * Updates the authenticated user's profile.
   * @param userData - An object containing the fields to update (e.g., username).
   * @returns An Observable of the updated user profile data.
   * @description Hits the `PATCH /api/v1/users/me` endpoint.
   */
  updateUser(userData: UpdateUserRequest): Observable<UpdateUserResponse> {
    return this.http.patch<UpdateUserResponse>(`${this.apiUrl}/me`, userData);
  }

  /**
   * Deletes the authenticated user's account.
   * @returns An empty Observable that completes on success.
   * @description Hits the `DELETE /api/v1/users/me` endpoint.
   */
  deleteUser(): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/me`);
  }

  /**
   * Fetches the public profile information for a specific user by ID.
   * @param id - The user's ID.
   * @returns An Observable of the user's public profile data.
   * @description Hits the `GET /api/v1/users/{id}` endpoint.
   */
  getUserProfile(id: string | number): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/${id}`);
  }

  /**
   * Fetches a paginated list of recent matches for a specific user.
   * @param id - The user's ID.
   * @param pageNum - Page number for pagination (default: 1).
   * @param limit - Number of results per page (default: 10, max: 50).
   * @returns An Observable of an array of Session objects.
   * @description Hits the `GET /api/v1/users/{id}/matches?page={page_num}&limit={limit}` endpoint.
   */
  getUserMatches(id: string | number, pageNum = 1, limit = 10): Observable<Session[]> {
    const params = new HttpParams()
      .set('page_num', pageNum.toString())
      .set('limit', limit.toString());

    return this.http.get<Session[]>(`${this.apiUrl}/${id}/matches`, { params });
  }
}
