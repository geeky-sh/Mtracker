import React from 'react';
import { Pressable, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Activity } from '@/types';
import { contrastColor } from '@/utils/colors';

const DAY_ABBR = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTH_ABBR = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function formatLastLogged(iso: string): string {
  const d = new Date(iso);
  return `${DAY_ABBR[d.getDay()]}, ${MONTH_ABBR[d.getMonth()]} ${d.getDate()}`;
}

interface Props {
  activity: Activity;
  onPress?: () => void;
  onDelete?: () => void;
  selected?: boolean;
  showDate?: boolean;
}

export function ActivityCard({ activity, onPress, onDelete, selected, showDate = true }: Props) {
  const textColor = contrastColor(activity.color);
  return (
    <Pressable
      onPress={onPress}
      style={[
        styles.card,
        { backgroundColor: activity.color },
        selected && styles.selected,
      ]}
    >
      <View style={styles.row}>
        <View style={styles.info}>
          <Text style={[styles.name, { color: textColor }]} numberOfLines={1}>
            {activity.name}
          </Text>
          {showDate && (
            <Text style={[styles.date, { color: textColor }]}>
              {activity.last_logged_date
                ? `Last: ${formatLastLogged(activity.last_logged_date)}`
                : 'Not tracked yet'}
            </Text>
          )}
          {activity.description ? (
            <Text style={[styles.desc, { color: textColor }]} numberOfLines={2}>
              {activity.description}
            </Text>
          ) : null}
        </View>
        {onDelete && (
          <TouchableOpacity onPress={onDelete} hitSlop={8} style={styles.deleteBtn}>
            <Ionicons name="trash-outline" size={18} color={textColor} style={{ opacity: 0.75 }} />
          </TouchableOpacity>
        )}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 12,
    padding: 14,
    marginBottom: 10,
  },
  selected: {
    borderWidth: 3,
    borderColor: '#1a1a1a',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  info: {
    flex: 1,
  },
  name: {
    fontSize: 16,
    fontWeight: '700',
  },
  date: {
    fontSize: 12,
    fontWeight: '500',
    marginTop: 3,
  },
  desc: {
    fontSize: 13,
    marginTop: 4,
    opacity: 0.85,
  },
  deleteBtn: {
    padding: 4,
    marginLeft: 8,
  },
});
