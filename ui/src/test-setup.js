/**
 * Phase 205 WS5 — global test setup for Vue Test Utils.
 * Registers app-wide directives that main.js wires in production.
 */
import { config } from '@vue/test-utils'
import { navHint } from './directives/navHint.js'

config.global.directives = {
  ...(config.global.directives || {}),
  'nav-hint': navHint,
}
