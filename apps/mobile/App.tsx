import 'react-native-url-polyfill/auto';
import React, { useEffect, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { LoginScreen } from '@/screens/LoginScreen';
import { AppNavigator } from '@/navigation/AppNavigator';
import { isAuthenticated } from '@/services/auth';
import { getProfile } from '@/services/api';
import { User } from '@/types';

export default function App() {
  const [user, setUser] = useState<User | null>(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    isAuthenticated()
      .then(async (authed) => {
        if (authed) {
          const profile = await getProfile().catch(() => null);
          setUser(profile);
        }
      })
      .finally(() => setChecking(false));
  }, []);

  if (checking) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#f8f9fa' }}>
        <ActivityIndicator size="large" color="#4ECDC4" />
      </View>
    );
  }

  return (
    <SafeAreaProvider>
      <StatusBar style="dark" />
      {user ? (
        <AppNavigator />
      ) : (
        <LoginScreen onLogin={(u) => setUser(u)} />
      )}
    </SafeAreaProvider>
  );
}
