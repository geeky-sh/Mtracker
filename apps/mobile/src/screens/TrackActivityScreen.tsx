import React, { useCallback, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { useFocusEffect, useNavigation, useRoute } from '@react-navigation/native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { createLog, listActivities } from '@/services/api';
import { Activity } from '@/types';
import { ActivityCard } from '@/components/ActivityCard';

const DAY_ABBR = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTH_ABBR = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

function buildDates(): Array<{ value: string; day: string; date: number; month: string; isToday: boolean }> {
  const today = new Date();
  return Array.from({ length: 31 }, (_, i) => {
    const d = new Date(today);
    d.setDate(d.getDate() - i);
    return {
      value: formatDate(d),
      day: DAY_ABBR[d.getDay()],
      date: d.getDate(),
      month: MONTH_ABBR[d.getMonth()],
      isToday: i === 0,
    };
  });
}

const DATES = buildDates();

export function TrackActivityScreen() {
  const insets = useSafeAreaInsets();
  const navigation = useNavigation<any>();
  const route = useRoute<any>();
  const [activities, setActivities] = useState<Activity[]>([]);
  const [selected, setSelected] = useState<Activity | null>(null);
  const [selectedDate, setSelectedDate] = useState<string>(DATES[0].value);
  const [loading, setLoading] = useState(true);
  const [logging, setLogging] = useState(false);
  const dateListRef = useRef<FlatList>(null);

  useFocusEffect(
    useCallback(() => {
      setLoading(true);
      const preSelectedId = route.params?.activityId;
      listActivities()
        .then((data) => {
          setActivities(data);
          if (preSelectedId) {
            const match = data.find((a) => a.id === preSelectedId);
            if (match) setSelected(match);
            navigation.setParams({ activityId: undefined });
          }
        })
        .catch(() => Alert.alert('Error', 'Failed to load activities.'))
        .finally(() => setLoading(false));
    }, [route.params?.activityId]),
  );

  const handleLog = async () => {
    if (!selected) {
      Alert.alert('Select an activity first.');
      return;
    }
    setLogging(true);
    try {
      await createLog(selected.id, selectedDate);
      Alert.alert('Logged!', `"${selected.name}" recorded for ${selectedDate}.`);
      setSelected(null);
    } catch (e: any) {
      const msg = e?.response?.data?.error ?? 'Failed to log activity.';
      Alert.alert('Error', msg);
    } finally {
      setLogging(false);
    }
  };

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <Text style={styles.heading}>Track Activity</Text>

      <Text style={styles.label}>Select date</Text>
      <FlatList
        ref={dateListRef}
        data={DATES}
        horizontal
        keyExtractor={(item) => item.value}
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.dateRow}
        renderItem={({ item }) => {
          const active = selectedDate === item.value;
          return (
            <TouchableOpacity
              style={[styles.dateChip, active && styles.dateChipActive]}
              onPress={() => setSelectedDate(item.value)}
            >
              <Text style={[styles.dateChipDay, active && styles.dateChipTextActive]}>
                {item.isToday ? 'Today' : item.day}
              </Text>
              <Text style={[styles.dateChipNum, active && styles.dateChipTextActive]}>
                {item.date}
              </Text>
              <Text style={[styles.dateChipMonth, active && styles.dateChipTextActive]}>
                {item.month}
              </Text>
            </TouchableOpacity>
          );
        }}
      />

      <Text style={styles.label}>Select activity</Text>

      {loading ? (
        <ActivityIndicator size="large" color="#4ECDC4" style={{ marginTop: 20 }} />
      ) : activities.length === 0 ? (
        <Text style={styles.empty}>No activities yet. Create one first!</Text>
      ) : (
        <FlatList
          data={activities}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <ActivityCard
              activity={item}
              selected={selected?.id === item.id}
              onPress={() => setSelected(selected?.id === item.id ? null : item)}
              showDate={false}
            />
          )}
          contentContainerStyle={styles.list}
          showsVerticalScrollIndicator={false}
        />
      )}

      {selected && (
        <View style={styles.footer}>
          <TouchableOpacity
            style={[styles.logButton, { backgroundColor: selected.color }]}
            onPress={handleLog}
            disabled={logging}
          >
            {logging ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={styles.logButtonText}>
                Log "{selected.name}" for {selectedDate}
              </Text>
            )}
          </TouchableOpacity>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
    padding: 16,
  },
  heading: {
    fontSize: 28,
    fontWeight: '800',
    color: '#1a1a1a',
    marginTop: 8,
    marginBottom: 20,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#555',
    marginBottom: 8,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  dateRow: {
    paddingBottom: 4,
    gap: 8,
    marginBottom: 20,
  },
  dateChip: {
    alignItems: 'center',
    borderWidth: 1.5,
    borderColor: '#ddd',
    borderRadius: 12,
    paddingVertical: 10,
    paddingHorizontal: 14,
    backgroundColor: '#fff',
    minWidth: 58,
  },
  dateChipActive: {
    borderColor: '#1a1a1a',
    backgroundColor: '#1a1a1a',
  },
  dateChipDay: {
    fontSize: 11,
    fontWeight: '600',
    color: '#888',
    textTransform: 'uppercase',
    letterSpacing: 0.3,
  },
  dateChipNum: {
    fontSize: 20,
    fontWeight: '800',
    color: '#1a1a1a',
    lineHeight: 26,
  },
  dateChipMonth: {
    fontSize: 11,
    color: '#888',
    fontWeight: '500',
  },
  dateChipTextActive: {
    color: '#fff',
  },
  list: {
    paddingBottom: 100,
  },
  empty: {
    color: '#aaa',
    textAlign: 'center',
    marginTop: 32,
    fontSize: 15,
  },
  footer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: 16,
    backgroundColor: '#f8f9fa',
    borderTopWidth: 1,
    borderTopColor: '#eee',
  },
  logButton: {
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  logButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '700',
  },
});
