import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { DEFAULT_THEME, ironCoreThemes } from '@/styles/theme'

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: DEFAULT_THEME,
    themes: ironCoreThemes
  },
  defaults: {
    VCard:         { rounded: 'md', elevation: 0 },
    VBtn:          { rounded: 'md' },
    VTextField:    { variant: 'outlined', density: 'compact', rounded: 'md' },
    VSelect:       { variant: 'outlined', density: 'compact', rounded: 'md' },
    VAutocomplete: { variant: 'outlined', density: 'compact', rounded: 'md' },
    VDataTable:    { density: 'compact' },
  }
})
