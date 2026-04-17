import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Zones from '../views/Zones.vue'
import ZoneDetail from '../views/ZoneDetail.vue'
import Sensors from '../views/Sensors.vue'
import SensorDetail from '../views/SensorDetail.vue'
import Actuators from '../views/Actuators.vue'
import Schedules from '../views/Schedules.vue'
import Tasks from '../views/Tasks.vue'
import Inventory from '../views/Inventory.vue'
import Costs from '../views/Costs.vue'
import Fertigation from '../views/Fertigation.vue'
import Alerts from '../views/Alerts.vue'
import Plants from '../views/Plants.vue'
import CommonsCatalog from '../views/CommonsCatalog.vue'
import Settings from '../views/Settings.vue'
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
  { path: '/tasks',        component: Tasks,        name: 'tasks' },
  { path: '/fertigation',  component: Fertigation,  name: 'fertigation' },
  { path: '/inventory',    component: Inventory,    name: 'inventory' },
  { path: '/costs',        component: Costs,        name: 'costs' },
  { path: '/alerts',       component: Alerts,       name: 'alerts' },
  { path: '/plants',       component: Plants,       name: 'plants' },
  { path: '/catalog',      component: CommonsCatalog, name: 'catalog' },
  { path: '/settings',     component: Settings,     name: 'settings' },
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
