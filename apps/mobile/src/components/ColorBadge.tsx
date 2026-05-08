import React from 'react';
import { StyleSheet, View } from 'react-native';

interface Props {
  color: string;
  size?: number;
}

export function ColorBadge({ color, size = 14 }: Props) {
  return (
    <View
      style={[
        styles.badge,
        { backgroundColor: color, width: size, height: size, borderRadius: size / 2 },
      ]}
    />
  );
}

const styles = StyleSheet.create({
  badge: {
    marginRight: 8,
  },
});
