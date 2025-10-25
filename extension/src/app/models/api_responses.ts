import { MatchDetails } from './match';

export interface UserInfoResponse {
  id: number;
  username: string;
  discriminator: string;
  lc_username: string;
  avatar_url: string;
  rating: number;
}

export interface InviteNotification {
  from_user: UserInfoResponse;
  matchDetails: MatchDetails;
  createdAt: string;
}

export interface NotificationsResponse {
  invites: InviteNotification[];
}