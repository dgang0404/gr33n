import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Zones from '../views/Zones.vue'
import ZoneDetail from '../views/ZoneDetail.vue'
import Sensors from '../views/Sensors.vue'
import SensorDetail from '../views/SensorDetail.vue'
import Actuators from '../views/Actuators.vue'
import Schedules from '../views/Schedules.vue'
import Automation from '../views/Automation.vue'
import Tasks from '../views/Tasks.vue'
import Inventory from '../views/Inventory.vue'
import Costs from '../views/Costs.vue'
import Fertigation from '../views/Fertigation.vue'
import Alerts from '../views/Alerts.vue'
import Plants from '../views/Plants.vue'
import Animals from '../views/Animals.vue'
import Aquaponics from '../views/Aquaponics.vue'
import CommonsCatalog from '../views/CommonsCatalog.vue'
import Setpoints from '../views/Setpoints.vue'
import Settings from '../views/Settings.vue'
import FarmKnowledge from '../views/FarmKnowledge.vue'
import FarmGuardianChat from '../views/FarmGuardianChat.vue'
import OperatorGuide from '../views/OperatorGuide.vue'
import CropCycleSummary from '../views/CropCycleSummary.vue'
import CropCycleCompare from '../views/CropCycleCompare.vue'
import LightingPrograms from '../views/LightingPrograms.vue'
import FeedingHub from '../views/FeedingHub.vue'
import ComfortTargetsHub from '../views/ComfortTargetsHub.vue'
import SuppliesHub from '../views/SuppliesHub.vue'
import FeedingAdminHub from '../views/FeedingAdminHub.vue'
import MoneyHub from '../views/MoneyHub.vue'
import FarmSetupWizard from '../views/FarmSetupWizard.vue'
import ZoneSetupWizard from '../views/ZoneSetupWizard.vue'
import Login from '../views/Login.vue'

const routes = [
  { path: '/login',        component: Login,        name: 'login',        meta: { public: true } },
  { path: '/register',     component: Login,        name: 'register',     meta: { public: true } },
  { path: '/',             component: Dashboard,    name: 'dashboard' },
  { path: '/zones',        component: Zones,        name: 'zones' },
  { path: '/zones/:id',    component: ZoneDetail,   name: 'zone-detail' },
  { path: '/sensors',      component: Sensors,      name: 'sensors' },
  { path: '/sensors/:id', component: SensorDetail, name: 'sensor-detail' },
  { path: '/actuators',    component: Actuators,    name: 'actuators' },
  { path: '/schedules',    component: Schedules,    name: 'schedules' },
  { path: '/automation',   component: Automation,   name: 'automation' },
  { path: '/setpoints',    component: Setpoints,    name: 'setpoints' },
  { path: '/tasks',        component: Tasks,        name: 'tasks' },
  { path: '/feeding',     component: FeedingHub,     name: 'feeding' },
  { path: '/comfort-targets', component: ComfortTargetsHub, name: 'comfort-targets' },
  { path: '/fertigation',  component: Fertigation,  name: 'fertigation' },
  { path: '/inventory',    component: Inventory,    name: 'inventory' },
  { path: '/costs',        component: Costs,        name: 'costs' },
  // Phase 43 — farmer operations hubs
  { path: '/operations/supplies', component: SuppliesHub, name: 'operations-supplies' },
  { path: '/operations/feeding', component: FeedingAdminHub, name: 'operations-feeding' },
  { path: '/operations/money', component: MoneyHub, name: 'operations-money' },
  { path: '/alerts',       component: Alerts,       name: 'alerts' },
  { path: '/plants',       component: Plants,       name: 'plants' },
  { path: '/animals',      component: Animals,      name: 'animals' },
  { path: '/aquaponics',   component: Aquaponics,   name: 'aquaponics' },
  { path: '/catalog',      component: CommonsCatalog, name: 'catalog' },
  { path: '/farm-knowledge', component: FarmKnowledge, name: 'farm-knowledge' },
  { path: '/chat',         component: FarmGuardianChat, name: 'farm-guardian-chat' },
  { path: '/guardian/requests', redirect: { path: '/chat', query: { tab: 'pending' } } },
  { path: '/settings',     component: Settings,     name: 'settings' },
  { path: '/operator-guide', component: OperatorGuide, name: 'operator-guide' },
  // Phase 28 WS2 — crop cycle analytics
  { path: '/crop-cycles/:id/summary', component: CropCycleSummary, name: 'crop-cycle-summary' },
  { path: '/farms/:fid/crop-cycles/compare', component: CropCycleCompare, name: 'crop-cycle-compare' },
  { path: '/farms/:id/setup', component: FarmSetupWizard, name: 'farm-setup' },
  { path: '/farms/:id/zones/new', component: ZoneSetupWizard, name: 'zone-setup' },
  { path: '/lighting', component: LightingPrograms, name: 'lighting' },
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
