import React, { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { login } from '@/services/auth';
import { User } from '@/types';

interface Props {
  onLogin: (user: User) => void;
}

export function LoginScreen({ onLogin }: Props) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleLogin() {
    if (!username || !password) {
      Alert.alert('Enter username and password.');
      return;
    }
    setLoading(true);
    try {
      const result = await login(username.trim(), password);
      onLogin(result.user);
    } catch {
      Alert.alert('Login failed', 'Invalid username or password.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={styles.container}>
      <Text style={styles.logo}>Mtracker</Text>
      <Text style={styles.tagline}>Track what matters, every day.</Text>

      <TextInput
        style={styles.input}
        placeholder="Username"
        placeholderTextColor="#999"
        autoCapitalize="none"
        value={username}
        onChangeText={setUsername}
      />
      <TextInput
        style={styles.input}
        placeholder="Password"
        placeholderTextColor="#999"
        secureTextEntry
        value={password}
        onChangeText={setPassword}
      />

      {loading ? (
        <ActivityIndicator size="large" color="#4ECDC4" style={styles.spinner} />
      ) : (
        <TouchableOpacity style={styles.button} onPress={handleLogin}>
          <Text style={styles.buttonText}>Sign In</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f8f9fa',
    padding: 32,
  },
  logo: {
    fontSize: 42,
    fontWeight: '800',
    color: '#1a1a1a',
    letterSpacing: -1,
  },
  tagline: {
    fontSize: 16,
    color: '#666',
    marginTop: 8,
    marginBottom: 48,
  },
  input: {
    width: '100%',
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 12,
    paddingVertical: 14,
    paddingHorizontal: 16,
    fontSize: 16,
    color: '#1a1a1a',
    marginBottom: 12,
  },
  button: {
    width: '100%',
    backgroundColor: '#1a1a1a',
    paddingVertical: 16,
    borderRadius: 12,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  spinner: {
    marginTop: 16,
  },
});
