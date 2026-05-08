import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import Svg, { G, Path } from 'react-native-svg';
import { ActivitySummary } from '@/types';

interface Props {
  data: ActivitySummary[];
  size?: number;
}

function polarToCartesian(cx: number, cy: number, r: number, angle: number) {
  return {
    x: cx + r * Math.sin(angle),
    y: cy - r * Math.cos(angle),
  };
}

function slicePath(cx: number, cy: number, r: number, start: number, end: number): string {
  const s = polarToCartesian(cx, cy, r, start);
  const e = polarToCartesian(cx, cy, r, end);
  const large = end - start > Math.PI ? 1 : 0;
  return `M ${cx} ${cy} L ${s.x.toFixed(3)} ${s.y.toFixed(3)} A ${r} ${r} 0 ${large} 1 ${e.x.toFixed(3)} ${e.y.toFixed(3)} Z`;
}

export function PieChart({ data, size = 220 }: Props) {
  const total = data.reduce((sum, d) => sum + d.count, 0);

  if (total === 0) {
    return <Text style={styles.empty}>No activity logged in this period.</Text>;
  }

  const cx = size / 2;
  const cy = size / 2;
  const r = size / 2 - 6;

  let currentAngle = 0;
  const slices = data
    .filter((d) => d.count > 0)
    .map((d) => {
      const angle = (d.count / total) * 2 * Math.PI;
      const start = currentAngle;
      currentAngle += angle;
      return { ...d, start, end: currentAngle };
    });

  return (
    <View style={styles.container}>
      <Svg width={size} height={size}>
        <G>
          {slices.map((s) => (
            <Path
              key={s.activity_id}
              d={slicePath(cx, cy, r, s.start, s.end)}
              fill={s.color}
              stroke="#f8f9fa"
              strokeWidth={2}
            />
          ))}
        </G>
      </Svg>

      <View style={styles.legend}>
        {data.map((d) => {
          const pct = total > 0 ? Math.round((d.count / total) * 100) : 0;
          return (
            <View key={d.activity_id} style={styles.legendRow}>
              <View style={[styles.dot, { backgroundColor: d.color }]} />
              <Text style={styles.legendName} numberOfLines={1}>{d.name}</Text>
              <Text style={styles.legendStat}>{d.count}x · {pct}%</Text>
            </View>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
  },
  empty: {
    color: '#aaa',
    textAlign: 'center',
    marginVertical: 24,
    fontSize: 14,
  },
  legend: {
    width: '100%',
    marginTop: 16,
    gap: 8,
  },
  legendRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  dot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    flexShrink: 0,
  },
  legendName: {
    flex: 1,
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  legendStat: {
    fontSize: 13,
    color: '#777',
    fontWeight: '500',
  },
});
