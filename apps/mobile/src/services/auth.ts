import * as Google from 'expo-auth-session/providers/google';
import * as WebBrowser from 'expo-web-browser';

import { clearToken, getToken, loginWithGoogle as apiLoginWithGoogle, loginWithPassword } from './api';

WebBrowser.maybeCompleteAuthSession();

export async function login(
  username: string,
  password: string,
): Promise<{ token: string; user: import('@/types').User }> {
  return loginWithPassword(username, password);
}

export function useGoogleAuth() {
  const [request, response, promptAsync] = Google.useAuthRequest({
    clientId: process.env.EXPO_PUBLIC_GOOGLE_CLIENT_ID,
    iosClientId: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
    androidClientId: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  });
  return { request, response, promptAsync };
}

export async function handleGoogleResponse(
  response: Google.AuthSessionResult | null,
): Promise<{ token: string; user: import('@/types').User } | null> {
  if (response?.type !== 'success') return null;
  const accessToken = response.authentication?.accessToken;
  if (!accessToken) return null;
  return apiLoginWithGoogle(accessToken);
}

export async function isAuthenticated(): Promise<boolean> {
  const token = await getToken();
  return !!token;
}

export async function logout() {
  await clearToken();
}
