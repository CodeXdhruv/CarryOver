import { writable } from 'svelte/store';
import type { Manifest } from './engine';

export type ScreenType = 'home' | 'scan' | 'backup' | 'restore' | 'tools' | 'settings';

export const currentScreen = writable<ScreenType>('home');
export const theme = writable<'auto' | 'light' | 'dark'>('dark');
export const backupManifest = writable<Manifest | null>(null);

export interface SystemInfo {
  hostname: string;
  os_id: string;
  os_name: string;
  os_version: string;
  arch: string;
  username: string;
  home: string;
  shell: string;
  package_manager: string;
  restic_available: boolean;
  restic_version?: string;
  disk_free_gb?: number;
}

export const systemInfo = writable<SystemInfo | null>(null);

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
