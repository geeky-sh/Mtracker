export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  created_at: string;
}

export interface Activity {
  id: string;
  user_id: string;
  name: string;
  description: string;
  color: string;
  created_at: string;
  last_logged_date?: string;
}

export interface ActivityLog {
  id: string;
  activity_id: string;
  user_id: string;
  logged_date: string; // ISO date string
  created_at: string;
}

export interface ActivitySummary {
  activity_id: string;
  name: string;
  color: string;
  count: number;
}

// react-native-calendars multi-dot marked dates shape
export type MarkedDates = Record<
  string,
  { dots: Array<{ key: string; color: string }> }
>;
