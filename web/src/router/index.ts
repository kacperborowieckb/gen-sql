import { createRouter, createWebHistory } from 'vue-router'

import GenerationView from '../views/GenerationView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: GenerationView,
    },
    {
      path: '/talk-to-data',
      name: 'talkToData',
      component: () => import('../views/TalkToDataView.vue'),
    },
  ],
})

export default router
