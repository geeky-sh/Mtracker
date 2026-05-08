import { useEffect } from 'react';

export function useDebounce<T>(
  value: T,
  delay: number,
  callback: (v: T) => void,
) {
  useEffect(() => {
    const timer = setTimeout(() => callback(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
}
