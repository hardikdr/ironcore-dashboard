import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'dark',
    themes: {
      light: {
        dark: false,
        colors: {
          primary:               '#6366f1',
          'primary-darken-1':    '#4f46e5',
          secondary:             '#8b5cf6',
          accent:                '#06b6d4',
          success:               '#10b981',
          warning:               '#f59e0b',
          error:                 '#ef4444',
          info:                  '#3b82f6',
          background:            '#f1f5f9',
          surface:               '#ffffff',
          'surface-variant':     '#f8fafc',
          'on-surface-variant':  '#64748b',
        }
      },
      dark: {
        dark: true,
        colors: {
          primary:               '#818cf8',
          'primary-darken-1':    '#6366f1',
          secondary:             '#a78bfa',
          accent:                '#22d3ee',
          success:               '#34d399',
          warning:               '#fbbf24',
          error:                 '#f87171',
          info:                  '#60a5fa',
          background:            '#070b14',
          surface:               '#0d1526',
          'surface-variant':     '#141e33',
          'on-surface-variant':  '#94a3b8',
        }
      }
    }
  },
  defaults: {
    VCard:         { rounded: 'lg', elevation: 0 },
    VBtn:          { rounded: 'lg' },
    VTextField:    { variant: 'outlined', density: 'compact', rounded: 'lg' },
    VSelect:       { variant: 'outlined', density: 'compact', rounded: 'lg' },
    VAutocomplete: { variant: 'outlined', density: 'compact', rounded: 'lg' },
    VDataTable:    { density: 'compact' },
  }
})
