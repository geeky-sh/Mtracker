import { clearToken, getToken, loginWithPassword } from './api';

export async function login(
  username: string,
  password: string,
): Promise<{ token: string; user: import('@/types').User }> {
  return loginWithPassword(username, password);
}

export async function isAuthenticated(): Promise<boolean> {
  const token = await getToken();
  return !!token;
}

export async function logout() {
  await clearToken();
}
