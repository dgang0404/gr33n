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
          label: 'My rooms',
          navTitle: 'Grow areas — water, light, and climate per room',
        },
        {
          to: '/feeding',
          icon: '💧',
          label: 'Feed & water',
          navTitle: 'Daily feeding plan per room — how each room gets water',
        },
        {
          to: '/comfort-targets',
          icon: '🎯',
          label: 'Targets & schedules',
          navTitle: 'Comfort bands, what runs when, and automation toggles',
        },
        { to: '/plants', icon: '🌱', label: 'Plants' },
        {
          to: '/inventory',
          icon: '🧪',
          label: 'Supplies',
          navTitle: 'Inventory and input batches',
        },
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
          navTitle: 'Sensor-triggered automation rules',
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
          label: 'Feeding (technical)',
          navTitle: 'Programs, reservoirs, EC targets, mixing log — power users',
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
        { to: '/costs', icon: '💰', label: 'Costs' },
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
        { to: '/operator-guide', icon: '📖', label: 'Guide' },
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

export const mobileBottomNav = [
  { to: '/', icon: '🌿', label: 'Today' },
  { to: '/zones', icon: '🗂️', label: 'Rooms' },
  { to: '/tasks', icon: '✅', label: 'Tasks' },
  { to: '/alerts', icon: '🔔', label: 'Alerts' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
