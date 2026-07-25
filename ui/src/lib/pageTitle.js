import { workspaceFor, WORKSPACES } from './workspaces.js'

/** Standalone routes (not workspace shells). */
const ROUTE_LABELS = {
  '/': 'Today',
  '/today': 'Today',
  '/virtual-pi': 'Wiring',
  '/pi-setup': 'Pi + HAT setup',
  '/chat': 'Farm Guardian',
  '/guardian/requests': 'Farm Guardian',
  '/sensors': 'Sensors',
  '/actuators': 'Controls',
  '/schedules': 'Schedules',
  '/tasks': 'Tasks',
  '/feeding': 'Feed & water',
  '/fertigation': 'Fertigation',
  '/inventory': 'Natural farming',
  '/alerts': 'Alerts',
  '/plants': 'Plants',
  '/catalog': 'Catalog',
  '/costs': 'Costs',
  '/settings': 'Settings',
  '/animals': 'Animals',
  '/aquaponics': 'Aquaponics',
  '/farm-knowledge': 'Farm knowledge',
  '/operator-guide': 'Operator guide',
}

/**
 * Primary chrome title for TopBar — workspace label when on a shell route.
 * @param {import('vue-router').RouteLocationNormalized} route
 */
export function pageTitleFromRoute(route) {
  const path = (route?.path ?? '').split('?')[0]
  if (!path) return 'gr33n'
  if (path.startsWith('/zones/')) return 'Zone detail'
  const wsHit = workspaceFor(path)
  if (wsHit?.workspaceId && WORKSPACES[wsHit.workspaceId]?.label) {
    return WORKSPACES[wsHit.workspaceId].label
  }
  return ROUTE_LABELS[path] ?? 'gr33n'
}
