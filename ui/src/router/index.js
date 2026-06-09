import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import ZonesWorkspace from '../views/workspaces/ZonesWorkspace.vue'
import ZoneDetail from '../views/ZoneDetail.vue'
import Sensors from '../views/Sensors.vue'
import SensorDetail from '../views/SensorDetail.vue'
import Actuators from '../views/Actuators.vue'
import Schedules from '../views/Schedules.vue'
import Automation from '../views/Automation.vue'
import Animals from '../views/Animals.vue'
import Aquaponics from '../views/Aquaponics.vue'
import CommonsCatalog from '../views/CommonsCatalog.vue'
import Setpoints from '../views/Setpoints.vue'
import Settings from '../views/Settings.vue'
import FarmKnowledge from '../views/FarmKnowledge.vue'
import FarmGuardianChat from '../views/FarmGuardianChat.vue'
import OperatorGuide from '../views/OperatorGuide.vue'
import CropProfileDetail from '../views/CropProfileDetail.vue'
import CropCycleSummary from '../views/CropCycleSummary.vue'
import CropCycleCompare from '../views/CropCycleCompare.vue'
import ComfortTargetsHub from '../views/ComfortTargetsHub.vue'
import FarmSetupWizard from '../views/FarmSetupWizard.vue'
import ZoneSetupWizard from '../views/ZoneSetupWizard.vue'
import DeviceSetupWizard from '../views/DeviceSetupWizard.vue'
import HardwareWorkspace from '../views/workspaces/HardwareWorkspace.vue'
import FeedWaterWorkspace from '../views/workspaces/FeedWaterWorkspace.vue'
import MoneyWorkspace from '../views/workspaces/MoneyWorkspace.vue'
import Login from '../views/Login.vue'
import { buildLegacyRedirectRoutes, buildZoneOpsRedirectRoutes } from '../lib/workspaces.js'

const routes = [
  { path: '/login',        component: Login,        name: 'login',        meta: { public: true } },
  { path: '/register',     component: Login,        name: 'register',     meta: { public: true } },
  { path: '/',             component: Dashboard,    name: 'dashboard' },
  { path: '/zones',        component: ZonesWorkspace, name: 'zones' },
  { path: '/zones/:id',    component: ZoneDetail,   name: 'zone-detail' },
  { path: '/hardware',     component: HardwareWorkspace, name: 'hardware' },
  { path: '/feed-water',   component: FeedWaterWorkspace, name: 'feed-water' },
  { path: '/money',        component: MoneyWorkspace, name: 'money' },
  { path: '/sensors/:id', component: SensorDetail, name: 'sensor-detail' },
  { path: '/schedules',    component: Schedules,    name: 'schedules' },
  { path: '/automation',   component: Automation,   name: 'automation' },
  { path: '/setpoints',    component: Setpoints,    name: 'setpoints' },
  { path: '/comfort-targets', component: ComfortTargetsHub, name: 'comfort-targets' },
  { path: '/crop-profiles/:id', component: CropProfileDetail, name: 'crop-profile-detail' },
  { path: '/animals',      component: Animals,      name: 'animals' },
  { path: '/aquaponics',   component: Aquaponics,   name: 'aquaponics' },
  { path: '/catalog',      component: CommonsCatalog, name: 'catalog' },
  { path: '/farm-knowledge', component: FarmKnowledge, name: 'farm-knowledge' },
  { path: '/chat',         component: FarmGuardianChat, name: 'farm-guardian-chat' },
  { path: '/guardian/requests', redirect: { path: '/chat', query: { tab: 'pending' } } },
  { path: '/settings',     component: Settings,     name: 'settings' },
  { path: '/operator-guide', component: OperatorGuide, name: 'operator-guide' },
  { path: '/crop-cycles/:id/summary', component: CropCycleSummary, name: 'crop-cycle-summary' },
  { path: '/farms/:fid/crop-cycles/compare', component: CropCycleCompare, name: 'crop-cycle-compare' },
  { path: '/farms/:id/setup', component: FarmSetupWizard, name: 'farm-setup' },
  { path: '/farms/:id/zones/new', component: ZoneSetupWizard, name: 'zone-setup' },
  { path: '/farms/:id/devices/new', component: DeviceSetupWizard, name: 'device-setup' },
  ...buildZoneOpsRedirectRoutes(),
  ...buildLegacyRedirectRoutes(),
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const token = localStorage.getItem('gr33n_token')
  if (!to.meta.public && !token) return { name: 'login' }
  if ((to.name === 'login' || to.name === 'register') && token) return { name: 'dashboard' }
})

export default router
