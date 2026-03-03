import { createRouter, createWebHistory } from 'vue-router'
import Dashboard   from '../views/Dashboard.vue'
import Zones       from '../views/Zones.vue'
import Sensors     from '../views/Sensors.vue'
import Actuators   from '../views/Actuators.vue'
import Schedules   from '../views/Schedules.vue'
import Inventory   from '../views/Inventory.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',           component: Dashboard,  name: 'dashboard'  },
    { path: '/zones',      component: Zones,      name: 'zones'      },
    { path: '/sensors',    component: Sensors,    name: 'sensors'    },
    { path: '/actuators',  component: Actuators,  name: 'actuators'  },
    { path: '/schedules',  component: Schedules,  name: 'schedules'  },
    { path: '/inventory',  component: Inventory,  name: 'inventory'  },
  ],
})
