export interface Profile {
  real_name?: string;
  student_id?: string;
  hobbies?: string;
  what_to_do?: string;
  comment?: string;
}

export interface User {
  id: string;
  discord_id: string;
  username: string;
  avatar_url: string;
  last_login_at?: string;
  profile?: Profile;
}

export interface MembersResponse {
  members: User[];
  total: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
}
