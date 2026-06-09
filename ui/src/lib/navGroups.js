import { GROW_PATH_ZONE_LABELS as Z } from './farmerVocabulary.js'

/**
 * Phase 68 — workspace-first sidebar navigation.
 * Workspaces replace scattered hub/admin/advanced duplicates.
 *
 * @param {string} cycleCompareRoute
 */
export function buildNavGroups(cycleCompareRoute) {
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
          navTitle: 'Spend summary, ledger, and supply costs',
        },
      ],
    },
    {
      label: 'More',
      items: [
        { to: '/animals', icon: '🐔', label: 'Animals' },
        { to: '/aquaponics', icon: '🐟', label: 'Aquaponics' },
        {
          to: cycleCompareRoute,
          icon: '📊',
          label: 'Analytics',
          navTitle: 'Crop cycle analytics',
        },
        { to: '/farm-knowledge', icon: '🔎', label: 'Knowledge' },
        {
          to: '/operator-guide',
          icon: '📖',
          label: 'Guide',
          navTitle: 'Operator glossary and recommended click paths',
        },
        { to: '/catalog', icon: '📚', label: 'Catalog' },
        {
          to: '/chat',
          icon: '💬',
          label: 'Guardian (full page)',
          navTitle: 'Full-page Farm Guardian chat and session history',
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
  { to: '/alerts', icon: '🔔', label: 'Alerts' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
