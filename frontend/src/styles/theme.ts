import type { ThemeDefinition } from 'vuetify'

export const DEFAULT_THEME = 'light' as const

export const ironCoreThemes: Record<'light' | 'dark', ThemeDefinition> = {
  light: {
    dark: false,
    colors: {
      primary: '#4f46e5',
      'primary-darken-1': '#4338ca',
      secondary: '#475569',
      success: '#16865b',
      warning: '#b45309',
      error: '#c2413b',
      info: '#2563eb',
      background: '#f6f7f9',
      surface: '#ffffff',
      'surface-variant': '#f1f3f6',
      'on-surface-variant': '#5f6b7a',
    },
  },
  dark: {
    dark: true,
    colors: {
      primary: '#a5b4fc',
      'primary-darken-1': '#818cf8',
      secondary: '#94a3b8',
      success: '#4ade9a',
      warning: '#fbbf24',
      error: '#fb7185',
      info: '#60a5fa',
      background: '#111318',
      surface: '#191c23',
      'surface-variant': '#222630',
      'on-surface-variant': '#aab2c0',
    },
  },
}
