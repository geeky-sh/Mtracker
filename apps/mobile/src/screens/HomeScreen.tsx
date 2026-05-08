import React, { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Modal,
  Pressable,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Calendar } from 'react-native-calendars';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { deleteActivity, deleteLog, getAnalytics, listActivities, listLogs } from '@/services/api';
import { Activity, ActivityLog, ActivitySummary, MarkedDates } from '@/types';
import { ActivityCard } from '@/components/ActivityCard';
import { PieChart } from '@/components/PieChart';

const DAY_OPTIONS = [15, 30, 90] as const;
type DayOption = typeof DAY_OPTIONS[number];

interface ActivityDropdownProps {
  items: { id: string; name: string; color: string }[];
  selectedIds: Set<string> | null;
  onToggle: (id: string) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
}

function ActivityDropdown({ items, selectedIds, onToggle, onSelectAll, onDeselectAll }: ActivityDropdownProps) {
  const [open, setOpen] = useState(false);
  const allSelected = selectedIds === null;
  const label = allSelected ? 'All Activities' : `${selectedIds.size} of ${items.length} selected`;

  if (items.length <= 1) return null;

  return (
    <View style={styles.dropdownWrapper}>
      <TouchableOpacity
        style={styles.dropdownTrigger}
        onPress={() => setOpen((o) => !o)}
        activeOpacity={0.8}
      >
        <Text style={styles.dropdownTriggerText}>{label}</Text>
        <Ionicons name={open ? 'chevron-up' : 'chevron-down'} size={16} color="#555" />
      </TouchableOpacity>

      {open && (
        <View style={styles.dropdownMenu}>
          <View style={styles.dropdownHeader}>
            <Text style={styles.dropdownHeaderLabel}>Filter Activities</Text>
            <TouchableOpacity onPress={allSelected ? onDeselectAll : onSelectAll}>
              <Text style={styles.dropdownHeaderAction}>
                {allSelected ? 'Deselect All' : 'Select All'}
              </Text>
            </TouchableOpacity>
          </View>

          {items.map((item) => {
            const checked = selectedIds === null || selectedIds.has(item.id);
            return (
              <TouchableOpacity
                key={item.id}
                style={styles.dropdownItem}
                onPress={() => onToggle(item.id)}
              >
                <View style={styles.dropdownItemLeft}>
                  <View style={[styles.colorDot, { backgroundColor: item.color }]} />
                  <Text style={styles.dropdownItemText} numberOfLines={1}>{item.name}</Text>
                </View>
                {checked && <Ionicons name="checkmark" size={16} color="#1a1a1a" />}
              </TouchableOpacity>
            );
          })}
        </View>
      )}
    </View>
  );
}

export function HomeScreen() {
  const insets = useSafeAreaInsets();
  const navigation = useNavigation<any>();

  const [activities, setActivities] = useState<Activity[]>([]);
  const [markedDates, setMarkedDates] = useState<MarkedDates>({});
  const [logsByDate, setLogsByDate] = useState<Record<string, ActivityLog[]>>({});
  const [loading, setLoading] = useState(true);
  const [logPickerModal, setLogPickerModal] = useState<{ day: string; logs: ActivityLog[] } | null>(null);

  const [analyticsDays, setAnalyticsDays] = useState<DayOption>(30);
  const [analyticsData, setAnalyticsData] = useState<ActivitySummary[]>([]);
  const [statsSelectedIds, setStatsSelectedIds] = useState<Set<string> | null>(null);
  const [loadingAnalytics, setLoadingAnalytics] = useState(false);

  const fetchAll = useCallback(async () => {
    try {
      const data = await listActivities();
      setActivities(data);

      if (data.length > 0) {
        const allLogs = await Promise.all(data.map((a) => listLogs(a.id)));
        const combined: MarkedDates = {};
        const byDate: Record<string, ActivityLog[]> = {};
        allLogs.forEach((logs, i) => {
          const activity = data[i];
          logs.forEach((log) => {
            const day = log.logged_date.split('T')[0];
            if (!combined[day]) combined[day] = { dots: [] };
            combined[day].dots.push({ key: activity.id, color: activity.color });
            if (!byDate[day]) byDate[day] = [];
            byDate[day].push(log);
          });
        });
        setMarkedDates(combined);
        setLogsByDate(byDate);
      } else {
        setMarkedDates({});
        setLogsByDate({});
      }
    } catch {
      Alert.alert('Error', 'Failed to load data.');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchAnalytics = useCallback(async (days: DayOption) => {
    setLoadingAnalytics(true);
    try {
      const data = await getAnalytics(days);
      setAnalyticsData(data);
    } catch {
      Alert.alert('Error', 'Failed to load analytics.');
    } finally {
      setLoadingAnalytics(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchAll();
      fetchAnalytics(analyticsDays);
    }, [fetchAll, fetchAnalytics, analyticsDays]),
  );

  const handleDayChange = (days: DayOption) => {
    setAnalyticsDays(days);
    fetchAnalytics(days);
  };

  const toggleStats = (id: string) => {
    setStatsSelectedIds((prev) => {
      const allIds = new Set(activities.map((a) => a.id));
      const current = prev ?? allIds;
      const next = new Set(current);
      if (next.has(id)) {
        if (next.size === 1) return prev;
        next.delete(id);
      } else {
        next.add(id);
      }
      return next.size === allIds.size ? null : next;
    });
  };

  const handleDeleteActivity = (activity: Activity) => {
    Alert.alert(
      'Delete Activity',
      `Delete "${activity.name}"? This will also delete all its logged history.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await deleteActivity(activity.id);
              setStatsSelectedIds((prev) => {
                if (prev === null) return null;
                const next = new Set(prev);
                next.delete(activity.id);
                return next.size === 0 ? null : next;
              });
              await Promise.all([fetchAll(), fetchAnalytics(analyticsDays)]);
            } catch {
              Alert.alert('Error', 'Failed to delete activity.');
            }
          },
        },
      ],
    );
  };

  const handleDayPress = (day: string) => {
    const logs = logsByDate[day];
    if (!logs || logs.length === 0) return;
    setLogPickerModal({ day, logs });
  };

  const handleDeleteLog = (log: ActivityLog, day: string) => {
    setLogPickerModal(null);
    const actName = activities.find((a) => a.id === log.activity_id)?.name ?? 'this activity';
    Alert.alert(
      'Confirm Delete',
      `Delete "${actName}" log for ${day}?`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await deleteLog(log.id);
              await fetchAll();
            } catch {
              Alert.alert('Error', 'Failed to delete log.');
            }
          },
        },
      ],
    );
  };

  const filteredMarkedDates: MarkedDates = statsSelectedIds === null
    ? markedDates
    : Object.fromEntries(
        Object.entries(markedDates)
          .map(([day, val]) => [
            day,
            { dots: val.dots.filter((dot) => statsSelectedIds.has(dot.key)) },
          ])
          .filter(([, val]) => (val as { dots: unknown[] }).dots.length > 0),
      );

  const visibleAnalytics = analyticsData.filter(
    (d) => statsSelectedIds === null || statsSelectedIds.has(d.activity_id),
  );

  const statsItems = activities.map((a) => ({ id: a.id, name: a.name, color: a.color }));

  const analyticsFooter = (
    <View>
      <Text style={styles.sectionTitle}>Stats</Text>

      <ActivityDropdown
        items={statsItems}
        selectedIds={statsSelectedIds}
        onToggle={toggleStats}
        onSelectAll={() => setStatsSelectedIds(null)}
        onDeselectAll={() => setStatsSelectedIds(new Set([activities[0]?.id].filter(Boolean)))}
      />

      <View style={styles.chipRow}>
        {DAY_OPTIONS.map((d) => (
          <TouchableOpacity
            key={d}
            style={[styles.chip, analyticsDays === d && styles.chipActive]}
            onPress={() => handleDayChange(d)}
          >
            <Text style={[styles.chipText, analyticsDays === d && styles.chipTextActive]}>
              {d}d
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      <Calendar
        markedDates={filteredMarkedDates}
        markingType="multi-dot"
        onDayPress={(d) => handleDayPress(d.dateString)}
        theme={{
          todayTextColor: '#4ECDC4',
          arrowColor: '#1a1a1a',
          textMonthFontWeight: '700',
        }}
        style={styles.calendar}
      />

      {loadingAnalytics ? (
        <ActivityIndicator size="large" color="#4ECDC4" style={{ marginVertical: 24 }} />
      ) : (
        <PieChart data={visibleAnalytics} />
      )}
    </View>
  );

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <Text style={styles.heading}>Home</Text>
      <Text style={styles.sectionTitle}>Your Activities</Text>

      {loading ? (
        <ActivityIndicator size="large" color="#4ECDC4" style={{ marginTop: 20 }} />
      ) : activities.length === 0 ? (
        <Text style={styles.empty}>No activities yet. Tap "New" to create one!</Text>
      ) : (
        <FlatList
          data={activities}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <ActivityCard
              activity={item}
              onPress={() => navigation.navigate('Track', { activityId: item.id })}
              onDelete={() => handleDeleteActivity(item)}
            />
          )}
          contentContainerStyle={styles.list}
          showsVerticalScrollIndicator={false}
          ListFooterComponent={analyticsFooter}
        />
      )}

      <Modal
        visible={logPickerModal !== null}
        transparent
        animationType="fade"
        onRequestClose={() => setLogPickerModal(null)}
      >
        <Pressable style={styles.modalBackdrop} onPress={() => setLogPickerModal(null)}>
          <Pressable style={styles.modalSheet} onPress={() => {}}>
            <Text style={styles.modalTitle}>Logs for {logPickerModal?.day}</Text>
            <Text style={styles.modalSubtitle}>Select a log to delete</Text>
            {logPickerModal?.logs.map((log) => {
              const name = activities.find((a) => a.id === log.activity_id)?.name ?? 'Unknown';
              return (
                <TouchableOpacity
                  key={log.id}
                  style={styles.modalItem}
                  onPress={() => handleDeleteLog(log, logPickerModal.day)}
                >
                  <Ionicons name="trash-outline" size={16} color="#e53935" style={{ marginRight: 10 }} />
                  <Text style={styles.modalItemText}>{name}</Text>
                </TouchableOpacity>
              );
            })}
            <TouchableOpacity style={styles.modalCancel} onPress={() => setLogPickerModal(null)}>
              <Text style={styles.modalCancelText}>Cancel</Text>
            </TouchableOpacity>
          </Pressable>
        </Pressable>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
    paddingHorizontal: 16,
  },
  heading: {
    fontSize: 28,
    fontWeight: '800',
    color: '#1a1a1a',
    marginTop: 16,
    marginBottom: 12,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#1a1a1a',
    marginBottom: 10,
    marginTop: 8,
  },
  list: {
    paddingBottom: 32,
  },
  calendar: {
    borderRadius: 12,
    elevation: 2,
    shadowColor: '#000',
    shadowOpacity: 0.06,
    shadowRadius: 8,
    marginBottom: 16,
  },
  empty: {
    color: '#aaa',
    textAlign: 'center',
    marginTop: 32,
    fontSize: 15,
  },
  chipRow: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 12,
  },
  chip: {
    borderWidth: 1.5,
    borderColor: '#ddd',
    borderRadius: 20,
    paddingVertical: 6,
    paddingHorizontal: 16,
    backgroundColor: '#fff',
  },
  chipActive: {
    borderColor: '#1a1a1a',
    backgroundColor: '#1a1a1a',
  },
  chipText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#555',
  },
  chipTextActive: {
    color: '#fff',
  },
  dropdownWrapper: {
    marginBottom: 12,
  },
  dropdownTrigger: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#fff',
    borderWidth: 1.5,
    borderColor: '#ddd',
    borderRadius: 10,
    paddingVertical: 10,
    paddingHorizontal: 14,
  },
  dropdownTriggerText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
  },
  dropdownMenu: {
    backgroundColor: '#fff',
    borderWidth: 1.5,
    borderColor: '#ddd',
    borderRadius: 10,
    marginTop: 4,
    overflow: 'hidden',
  },
  dropdownHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 9,
    paddingHorizontal: 14,
    backgroundColor: '#f5f5f5',
    borderBottomWidth: 1,
    borderBottomColor: '#e8e8e8',
  },
  dropdownHeaderLabel: {
    fontSize: 12,
    fontWeight: '700',
    color: '#888',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  dropdownHeaderAction: {
    fontSize: 13,
    fontWeight: '700',
    color: '#4ECDC4',
  },
  dropdownItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 11,
    paddingHorizontal: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  dropdownItemLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    flex: 1,
  },
  colorDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    flexShrink: 0,
  },
  dropdownItemText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
    flex: 1,
  },
  modalBackdrop: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.45)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  modalSheet: {
    backgroundColor: '#fff',
    borderRadius: 16,
    width: '100%',
    paddingVertical: 20,
    paddingHorizontal: 20,
  },
  modalTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#1a1a1a',
    marginBottom: 4,
  },
  modalSubtitle: {
    fontSize: 13,
    color: '#888',
    marginBottom: 16,
  },
  modalItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    borderTopWidth: 1,
    borderTopColor: '#f0f0f0',
  },
  modalItemText: {
    fontSize: 15,
    fontWeight: '500',
    color: '#e53935',
    flex: 1,
  },
  modalCancel: {
    marginTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#f0f0f0',
    paddingTop: 14,
    alignItems: 'center',
  },
  modalCancelText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#555',
  },
});
