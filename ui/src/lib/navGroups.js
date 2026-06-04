/**
 * Phase 38 — shared sidebar / mobile drawer navigation groups.
 * @param {string} cycleCompareRoute
 */
export function buildNavGroups(cycleCompareRoute) {
  return [
    {
      label: 'Grow',
      items: [
        { to: '/zones', icon: '🗂️', label: 'Zones', navTitle: 'Grow areas — open a zone for water, light, and climate in one place (Phase 38)' },
        { to: '/fertigation', icon: '💧', label: 'Fertigation', navTitle: 'Feeding programs and reservoirs' },
        { to: '/plants', icon: '🌱', label: 'Plants' },
        { to: '/inventory', icon: '🧪', label: 'Inventory' },
      ],
    },
    {
      label: 'Operate',
      items: [
        { to: '/', icon: '🌿', label: 'Dashboard' },
        { to: '/tasks', icon: '✅', label: 'Tasks' },
        { to: '/schedules', icon: '📅', label: 'Schedules', navTitle: 'Time-based automation (cron)' },
        { to: '/lighting', icon: '💡', label: 'Lighting', navTitle: 'Photoperiod programs (Phase 35)' },
      ],
    },
    {
      label: 'Advanced',
      items: [
        { to: '/automation', icon: '🤖', label: 'Rules', navTitle: 'Sensor-triggered automation rules' },
        { to: '/setpoints', icon: '🎯', label: 'Setpoints', navTitle: 'Target ranges per sensor type and growth stage' },
        { to: '/actuators', icon: '⚡', label: 'Controls', navTitle: 'All actuators — manual on/off and timed pulses' },
        { to: '/sensors', icon: '📡', label: 'Sensors', navTitle: 'All sensors and live readings' },
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
        { to: '/alerts', icon: '🔔', label: 'Alerts' },
        { to: '/costs', icon: '💰', label: 'Costs' },
        { to: cycleCompareRoute, icon: '📊', label: 'Analytics', navTitle: 'Crop cycle analytics' },
        { to: '/farm-knowledge', icon: '🔎', label: 'Knowledge' },
      ],
    },
    {
      label: 'System',
      items: [
        { to: '/operator-guide', icon: '📖', label: 'Guide' },
        { to: '/catalog', icon: '📚', label: 'Catalog' },
        { to: '/settings', icon: '⚙️', label: 'Settings' },
      ],
    },
  ]
}

export const mobileBottomNav = [
  { to: '/', icon: '🌿', label: 'Home' },
  { to: '/zones', icon: '🗂️', label: 'Zones' },
  { to: '/tasks', icon: '✅', label: 'Tasks' },
  { to: '/alerts', icon: '🔔', label: 'Alerts' },
  { to: '/settings', icon: '⚙️', label: 'More' },
]
