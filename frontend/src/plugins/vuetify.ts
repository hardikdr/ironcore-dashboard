import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary:    '#1a5fa8',
          secondary:  '#475569',
          success:    '#16a34a',
          warning:    '#d97706',
          error:      '#b91c1c',
          info:       '#0369a1',
          background: '#f4f6fa',
          surface:    '#ffffff',
        }
      }
    }
  }
})
