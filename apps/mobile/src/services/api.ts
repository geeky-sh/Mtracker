import axios from 'axios';
import * as SecureStore from 'expo-secure-store';

const TOKEN_KEY = 'mtracker_jwt';

const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://10.0.2.2:8080';

export const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  timeout: 10_000,
  headers: { 'Content-Type': 'application/json' },
});

// Attach stored JWT to every request
api.interceptors.request.use(async (config) => {
  const token = await SecureStore.getItemAsync(TOKEN_KEY);
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  console.log(`[API REQ] ${config.method?.toUpperCase()} ${config.baseURL}${config.url}`, {
    params: config.params,
    data: config.data,
  });
  return config;
});

api.interceptors.response.use(
  (response) => {
    console.log(`[API RES] ${response.status} ${response.config.url}`, response.data);
    return response;
  },
  (error) => {
    console.log(
      `[API ERR] ${error.response?.status ?? 'NO_RESPONSE'} ${error.config?.url}`,
      error.response?.data ?? error.message,
    );
    return Promise.reject(error);
  },
);

export async function saveToken(token: string) {
  await SecureStore.setItemAsync(TOKEN_KEY, token);
}

export async function getToken(): Promise<string | null> {
  return SecureStore.getItemAsync(TOKEN_KEY);
}

export async function clearToken() {
  await SecureStore.deleteItemAsync(TOKEN_KEY);
}

// ── Auth ─────────────────────────────────────────────────────────────────────

export async function loginWithPassword(username: string, password: string) {
  const { data } = await api.post('/auth/login', { username, password });
  await saveToken(data.token);
  return data as { token: string; user: import('@/types').User };
}

export async function getProfile() {
  const { data } = await api.get('/profile');
  return data as import('@/types').User;
}

// ── Activities ────────────────────────────────────────────────────────────────

export async function listActivities() {
  const { data } = await api.get('/activities');
  return data as import('@/types').Activity[];
}

export async function createActivity(name: string, description: string) {
  const { data } = await api.post('/activities', { name, description });
  return data as import('@/types').Activity;
}

export async function searchActivities(q: string) {
  const { data } = await api.get('/activities/search', { params: { q } });
  return data as import('@/types').Activity[];
}

export async function deleteActivity(id: string) {
  await api.delete(`/activities/${id}`);
}

// ── Logs ──────────────────────────────────────────────────────────────────────

export async function createLog(activityId: string, loggedDate: string) {
  const { data } = await api.post('/logs', {
    activity_id: activityId,
    logged_date: loggedDate,
  });
  return data as import('@/types').ActivityLog;
}

export async function listLogs(activityId: string) {
  const { data } = await api.get('/logs', { params: { activity_id: activityId } });
  return data as import('@/types').ActivityLog[];
}

export async function deleteLog(logId: string) {
  await api.delete(`/logs/${logId}`);
}

// ── Analytics ─────────────────────────────────────────────────────────────────

export async function getAnalytics(days: number) {
  const { data } = await api.get('/analytics', { params: { days } });
  return data as import('@/types').ActivitySummary[];
}
