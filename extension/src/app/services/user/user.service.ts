import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'environments/environment';
import { User } from 'app/models/user.model';

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
}
