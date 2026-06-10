import { GROW_PATH_ZONE_LABELS as Z } from './farmerVocabulary.js'

/**
 * Phase 68 / 77 — workspace-first sidebar navigation.
 * Analytics, Guardian full page, and reference pages live in workspaces — not More.
 */
export function buildNavGroups() {
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
          navTitle: 'Every room — grows, fleet sensors, controls, and lighting',
        },
        {
          to: '/feed-water',
          icon: '💧',
          label: 'Feed & water',
          navTitle: 'Daily watering, programs, nutrients, and advanced fertigation',
        },
        {
          to: '/comfort-targets',
          icon: '🎯',
          label: 'Comfort & automation',
          navTitle: 'Comfort bands, what runs when, automations, and raw setpoints',
        },
        {
          to: '/hardware',
          icon: '🔌',
          label: 'Hardware',
          navTitle: 'Pi devices, GPIO wiring, relay channels, and setup guide',
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
        { to: '/animals', icon: '🐔', label: 'Animals' },
        { to: '/aquaponics', icon: '🐟', label: 'Aquaponics' },
        {
          to: '/operator-guide',
          icon: '📖',
          label: 'Help',
          navTitle: 'Operator guide, farm knowledge search, and commons catalog',
        },
        { to: '/settings', icon: '⚙️', label: 'Settings' },
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
  { to: '/feed-water', icon: '💧', label: 'Feed' },
  { to: '/comfort-targets', icon: '🎯', label: 'Targets' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
