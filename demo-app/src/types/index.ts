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
  display_name: string;
  avatar_url: string;
  last_login_at?: string;
  joined_at?: string;
  guild_nickname?: string;
  guild_roles?: string[];
  profile?: Profile;
}

export interface MembersResponse {
  members: User[];
  limit: number;
  offset: number;
  count: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
}