/**
 * Phase 32 WS1 — map Vue router location to Guardian context_ref route anchor.
 * Sent on grounded chat turns when no Ask Guardian entity ref is active.
 */

const ROUTE_LABELS = {
  '/': 'Dashboard',
  '/zones': 'Zones',
  '/sensors': 'Sensors',
  '/actuators': 'Actuators',
  '/schedules': 'Schedules',
  '/automation': 'Automation',
  '/feeding': 'Feed & water',
  '/operations/supplies': 'Supplies',
  '/operations/feeding': 'Feeding (details)',
  '/operations/money': 'Money',
  '/comfort-targets': 'Targets & schedules',
  '/fertigation': 'Feeding (technical)',
  '/setpoints': 'Setpoints (raw)',
  '/tasks': 'Tasks',
  '/inventory': 'Supplies (full editor)',
  '/costs': 'Money (full editor)',
  '/alerts': 'Alerts',
  '/plants': 'Plants',
  '/animals': 'Animals',
  '/aquaponics': 'Aquaponics',
  '/catalog': 'Commons catalog',
  '/farm-knowledge': 'Farm knowledge',
  '/chat': 'Farm Guardian chat',
  '/guardian/requests': 'Farm Guardian',
  '/settings': 'Settings',
  '/operator-guide': 'Operator guide',
  '/pi-setup': 'Pi + HAT setup guide',
}

function labelFromPath(path) {
  if (ROUTE_LABELS[path]) return ROUTE_LABELS[path]
  if (path.startsWith('/zones/')) return 'Zone detail'
  if (path.startsWith('/sensors/')) return 'Sensor detail'
  if (path.includes('/crop-cycles/') && path.endsWith('/summary')) return 'Crop cycle summary'
  if (path.includes('/crop-cycles/compare')) return 'Crop cycle compare'
  if (path.startsWith('/farms/') && path.endsWith('/setup')) return 'Farm setup'
  if (path.startsWith('/farms/') && path.endsWith('/zones/new')) return 'Add zone'
  if (path.startsWith('/farms/') && path.endsWith('/devices/new')) return 'Connect edge device'
  return path
}

/** @param {import('vue-router').RouteLocationNormalized} route */
export function routeContextRefFromRoute(route) {
  if (!route?.path || route.meta?.public) return null
  const path = route.path.split('?')[0]
  if (!path || path === '/login' || path === '/register') return null
  return {
    type: 'route',
    path,
    name: labelFromPath(path),
  }
}
