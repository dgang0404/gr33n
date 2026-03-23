import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Zones from '../views/Zones.vue'
import Sensors from '../views/Sensors.vue'
import Actuators from '../views/Actuators.vue'
import Schedules from '../views/Schedules.vue'
import Inventory from '../views/Inventory.vue'
import Settings from '../views/Settings.vue'
import Login from '../views/Login.vue'

const routes = [
  { path: '/login',    component: Login,     name: 'login',     meta: { public: true } },
  { path: '/',         component: Dashboard, name: 'dashboard' },
  { path: '/zones',    component: Zones,     name: 'zones' },
  { path: '/sensors',  component: Sensors,   name: 'sensors' },
  { path: '/actuators',component: Actuators, name: 'actuators' },
  { path: '/schedules',component: Schedules, name: 'schedules' },
  { path: '/inventory',component: Inventory, name: 'inventory' },
  { path: '/settings', component: Settings,  name: 'settings' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const token = localStorage.getItem('gr33n_token')
  if (!to.meta.public && !token) return { name: 'login' }
  if (to.name === 'login' && token) return { name: 'dashboard' }
})

export default router
