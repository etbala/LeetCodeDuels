export interface User {
  id: number;
  username: string;
  discriminator: string;
  lc_username: string;
  avatar_url: string;
  rating: number;
}