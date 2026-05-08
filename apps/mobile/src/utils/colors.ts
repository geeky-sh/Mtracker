export const palette = [
  '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4',
  '#F7DC6F', '#DDA0DD', '#98D8C8', '#F0B27A',
  '#BB8FCE', '#85C1E9', '#82E0AA', '#F1948A',
];

/** Returns a readable text colour (black or white) for a given background hex. */
export function contrastColor(hex: string): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance > 0.6 ? '#1a1a1a' : '#ffffff';
}
