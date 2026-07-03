import { GROW_PATH_ZONE_LABELS as Z } from './farmerVocabulary.js'
import { isModuleEnabled, MODULE_SCHEMA } from './farmModules.js'

/**
 * Phase 68 / 77 / 78 — workspace-first sidebar navigation.
 * Phase 115 — optional modules map gates Animals / Aquaponics links.
 */
export function buildNavGroups(opts = {}) {
  const modules = opts.modules ?? {}
  const items = []

  if (isModuleEnabled(modules, MODULE_SCHEMA.animals)) {
    items.push({ to: '/animals', icon: '🐔', label: 'Animals' })
  }
  if (isModuleEnabled(modules, MODULE_SCHEMA.aquaponics)) {
    items.push({ to: '/aquaponics', icon: '🐟', label: 'Aquaponics' })
  }

  items.push(
    {
      to: '/operator-guide',
      icon: '📖',
      label: 'Help',
      navTitle: 'Operator guide, farm knowledge search, and commons catalog',
    },
    { to: '/settings', icon: '⚙️', label: 'Settings' },
  )

  return [
    {
      label: 'Today',
      items: [
        {
          to: '/',
          icon: '🌿',
          label: 'Today',
          navTitle: 'Morning dashboard — tasks, alerts, schedules',
        },
      ],
    },
    {
      label: 'Grow & operate',
      items: [
        {
          to: '/zones',
          icon: '🗂️',
          label: Z.navMyZones,
          navTitle: 'My zones, farm-wide hardware, sensors, controls, and lighting',
        },
        {
          to: '/comfort-targets',
          icon: '🎯',
          label: 'Comfort & automation',
          navTitle: 'Comfort bands, what runs when, automations, and raw setpoints',
        },
        {
          to: '/virtual-pi',
          icon: '🔌',
          label: 'Wiring',
          navTitle: 'Virtual Pi — see what\'s wired to every pin',
        },
        {
          to: '/money',
          icon: '💰',
          label: 'Money',
          navTitle: 'Spend summary, ledger, supply costs, and grow economics',
        },
      ],
    },
    {
      label: 'More',
      items: [
        {
          to: '/chat',
          icon: '✨',
          label: 'Farm Guardian',
          navTitle: 'Full-page chat, session history, and pending requests',
        },
        ...items,
      ],
    },
  ]
}

/** All sidebar `to` paths (flat). */
export function collectSidebarRoutes(groups) {
  return groups.flatMap((g) => g.items.map((i) => i.to))
}

export const mobileBottomNav = [
  { to: '/', icon: '🌿', label: 'Today' },
  { to: '/zones', icon: '🗂️', label: Z.mobileZones },
  { to: '/comfort-targets', icon: '🎯', label: 'Targets' },
  { to: '/money', icon: '💰', label: 'Money' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
