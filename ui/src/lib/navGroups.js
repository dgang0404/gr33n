import { GROW_PATH_ZONE_LABELS as Z } from './farmerVocabulary.js'

/**
 * Phase 38 + Phase 40 WS7 — sidebar / mobile drawer navigation groups.
 * Farmer-first labels; Advanced routes grouped for power users.
 *
 * @param {string} cycleCompareRoute
 */
export function buildNavGroups(cycleCompareRoute) {
  return [
    {
      label: 'Grow',
      items: [
        {
          to: '/zones',
          icon: '🗂️',
          label: Z.navMyZones,
          navTitle: Z.navMyZonesTitle,
        },
        {
          to: '/feeding',
          icon: '💧',
          label: 'Feed & water',
          navTitle: Z.feedingNavTitle,
        },
        {
          to: '/comfort-targets',
          icon: '🎯',
          label: 'Targets & schedules',
          navTitle: 'Comfort bands, what runs when, and automation toggles',
        },
        { to: '/plants', icon: '🌱', label: 'Plants' },
      ],
    },
    {
      label: 'Today',
      items: [
        {
          to: '/',
          icon: '🌿',
          label: 'Today',
          navTitle: 'Morning dashboard — tasks, alerts, schedules',
        },
        { to: '/tasks', icon: '✅', label: 'Tasks' },
        { to: '/alerts', icon: '🔔', label: 'Alerts' },
      ],
    },
    {
      label: 'Operations',
      items: [
        {
          to: '/operations/supplies',
          icon: '🧪',
          label: 'Supplies',
          navTitle: 'What is on hand, what is running low, and mixing recipes',
        },
        {
          to: '/operations/feeding',
          icon: '💦',
          label: 'Feeding admin',
          navTitle: 'Programs, reservoirs, EC targets, and mixing log — farm-wide admin',
        },
        {
          to: '/operations/money',
          icon: '💰',
          label: 'Money',
          navTitle: 'Spend summary, receipts, and cost attachments',
        },
      ],
    },
    {
      label: 'Operate',
      items: [
        {
          to: '/lighting',
          icon: '💡',
          label: 'Lighting',
          navTitle: 'Photoperiod programs',
        },
      ],
    },
    {
      label: 'Advanced',
      items: [
        {
          to: '/schedules',
          icon: '📅',
          label: 'Schedules (cron)',
          navTitle: 'Cron schedule editor — power users',
        },
        {
          to: '/automation',
          icon: '🤖',
          label: 'Automations',
          navTitle: 'Sensor-triggered automations',
        },
        {
          to: '/setpoints',
          icon: '🎯',
          label: 'Setpoints (raw)',
          navTitle: 'Farm-wide target ranges per sensor and stage',
        },
        {
          to: '/fertigation',
          icon: '💦',
          label: 'Fertigation',
          navTitle: 'Fertigation console — programs, reservoirs, EC targets, mixing log',
        },
        {
          to: '/actuators',
          icon: '⚡',
          label: 'Controls',
          navTitle: 'All actuators — manual on/off and timed pulses',
        },
        {
          to: '/sensors',
          icon: '📡',
          label: 'Sensors',
          navTitle: 'All sensors and live readings',
        },
      ],
    },
    {
      label: 'Livestock',
      items: [
        { to: '/animals', icon: '🐔', label: 'Animals' },
        { to: '/aquaponics', icon: '🐟', label: 'Aquaponics' },
      ],
    },
    {
      label: 'Monitor',
      items: [
        {
          to: cycleCompareRoute,
          icon: '📊',
          label: 'Analytics',
          navTitle: 'Crop cycle analytics',
        },
        { to: '/farm-knowledge', icon: '🔎', label: 'Knowledge' },
      ],
    },
    {
      label: 'System',
      items: [
        {
          to: '/operator-guide',
          icon: '📖',
          label: 'Guide',
          children: [
            {
              to: '/pi-setup',
              icon: '🔌',
              label: 'Pi + HAT setup',
              navTitle: 'Raspberry Pi + Sequent Microsystems relay HAT wiring guide',
            },
          ],
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

/** All sidebar `to` paths, including nested Guide children (e.g. `/pi-setup`). */
export function collectSidebarRoutes(groups) {
  return groups.flatMap((g) =>
    g.items.flatMap((i) => [i.to, ...(i.children || []).map((c) => c.to)]),
  )
}

export const mobileBottomNav = [
  { to: '/', icon: '🌿', label: 'Today' },
  { to: '/zones', icon: '🗂️', label: Z.mobileZones },
  { to: '/tasks', icon: '✅', label: 'Tasks' },
  { to: '/alerts', icon: '🔔', label: 'Alerts' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
