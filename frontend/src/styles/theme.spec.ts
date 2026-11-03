import { describe, expect, it } from 'vitest'
import { DEFAULT_THEME, ironCoreThemes } from './theme'
import tokens from './tokens.css?raw'

describe('design system theme', () => {
  it('is light-first with distinct semantic colors', () => {
    expect(DEFAULT_THEME).toBe('light')
    expect(ironCoreThemes.light.dark).toBe(false)
    expect(ironCoreThemes.dark.dark).toBe(true)
    expect(ironCoreThemes.light.colors?.primary).toBeDefined()
    expect(ironCoreThemes.light.colors?.success).not.toBe(ironCoreThemes.light.colors?.primary)
  })
  it('defines the semantic token contract', () => {
    for (const token of ['--ic-canvas','--ic-surface','--ic-border','--ic-focus','--ic-radius-md','--ic-space-8']) expect(tokens).toContain(token)
  })
})
