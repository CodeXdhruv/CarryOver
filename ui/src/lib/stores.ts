import { writable } from 'svelte/store';
import type { Manifest } from './engine';

export type ScreenType = 'home' | 'backup' | 'restore' | 'settings';

export const currentScreen = writable<ScreenType>('home');
export const theme = writable<'auto' | 'light' | 'dark'>('auto');
export const backupManifest = writable<Manifest | null>(null);

export interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'success';
  message: string;
}

export const activityLogs = writable<LogEntry[]>([]);

export function addLog(message: string, level: 'info' | 'warn' | 'error' | 'success' = 'info') {
  const entry: LogEntry = {
    timestamp: new Date().toLocaleTimeString(),
    level,
    message,
  };
  activityLogs.update((logs) => [entry, ...logs.slice(0, 99)]);
}
