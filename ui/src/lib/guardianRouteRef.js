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
  '/setpoints': 'Setpoints',
  '/tasks': 'Tasks',
  '/fertigation': 'Fertigation',
  '/inventory': 'Inventory',
  '/costs': 'Costs',
  '/alerts': 'Alerts',
  '/plants': 'Plants',
  '/animals': 'Animals',
  '/aquaponics': 'Aquaponics',
  '/catalog': 'Commons catalog',
  '/farm-knowledge': 'Farm knowledge',
  '/chat': 'Farm Guardian chat',
  '/guardian/requests': 'Guardian change requests',
  '/settings': 'Settings',
  '/operator-guide': 'Operator guide',
}

function labelFromPath(path) {
  if (ROUTE_LABELS[path]) return ROUTE_LABELS[path]
  if (path.startsWith('/zones/')) return 'Zone detail'
  if (path.startsWith('/sensors/')) return 'Sensor detail'
  if (path.includes('/crop-cycles/') && path.endsWith('/summary')) return 'Crop cycle summary'
  if (path.includes('/crop-cycles/compare')) return 'Crop cycle compare'
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
