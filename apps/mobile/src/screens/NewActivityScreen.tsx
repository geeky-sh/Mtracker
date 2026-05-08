import React, { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { createActivity, searchActivities } from '@/services/api';
import { Activity } from '@/types';
import { ColorBadge } from '@/components/ColorBadge';
import { useDebounce } from '@/utils/useDebounce';

export function NewActivityScreen() {
  const insets = useSafeAreaInsets();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [suggestions, setSuggestions] = useState<Activity[]>([]);
  const [creating, setCreating] = useState(false);
  const [success, setSuccess] = useState(false);

  const fetchSuggestions = useCallback(async (q: string) => {
    if (q.trim().length < 2) { setSuggestions([]); return; }
    const results = await searchActivities(q).catch(() => []);
    setSuggestions(results);
  }, []);

  useDebounce(name, 300, fetchSuggestions);

  const handleCreate = async () => {
    if (!name.trim()) {
      Alert.alert('Please enter an activity name.');
      return;
    }
    setCreating(true);
    try {
      await createActivity(name.trim(), description.trim());
      setName('');
      setDescription('');
      setSuggestions([]);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 2000);
    } catch (e: any) {
      Alert.alert('Error', e?.response?.data?.error ?? 'Failed to create activity.');
    } finally {
      setCreating(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={[styles.container, { paddingTop: insets.top }]}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <Text style={styles.heading}>New Activity</Text>

      <Text style={styles.label}>Name *</Text>
      <TextInput
        style={styles.input}
        placeholder="e.g. Morning run, Read book…"
        placeholderTextColor="#bbb"
        value={name}
        onChangeText={setName}
        maxLength={100}
        returnKeyType="next"
      />

      {suggestions.length > 0 && (
        <View style={styles.suggestionsBox}>
          <Text style={styles.suggestionsTitle}>Similar activities you already have:</Text>
          {suggestions.map((s) => (
            <View key={s.id} style={styles.suggestionRow}>
              <ColorBadge color={s.color} size={10} />
              <Text style={styles.suggestionText}>{s.name}</Text>
            </View>
          ))}
        </View>
      )}

      <Text style={styles.label}>Description (optional)</Text>
      <TextInput
        style={[styles.input, styles.textArea]}
        placeholder="A short note about this activity…"
        placeholderTextColor="#bbb"
        value={description}
        onChangeText={setDescription}
        maxLength={500}
        multiline
        numberOfLines={3}
        textAlignVertical="top"
      />

      {success && (
        <Text style={styles.successMsg}>Activity created!</Text>
      )}

      <TouchableOpacity style={styles.button} onPress={handleCreate} disabled={creating}>
        {creating ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <Text style={styles.buttonText}>Create Activity</Text>
        )}
      </TouchableOpacity>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
    padding: 20,
  },
  heading: {
    fontSize: 28,
    fontWeight: '800',
    color: '#1a1a1a',
    marginTop: 8,
    marginBottom: 24,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#555',
    marginBottom: 6,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  input: {
    backgroundColor: '#fff',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#e0e0e0',
    padding: 14,
    fontSize: 16,
    color: '#1a1a1a',
    marginBottom: 16,
  },
  textArea: {
    minHeight: 90,
  },
  suggestionsBox: {
    backgroundColor: '#fff8e1',
    borderRadius: 10,
    padding: 12,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#ffe082',
  },
  suggestionsTitle: {
    fontSize: 12,
    fontWeight: '700',
    color: '#f57f17',
    marginBottom: 8,
  },
  suggestionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  suggestionText: {
    fontSize: 14,
    color: '#555',
  },
  successMsg: {
    color: '#2e7d32',
    fontWeight: '700',
    textAlign: 'center',
    marginBottom: 12,
    fontSize: 15,
  },
  button: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
});
