import { createRouter, createWebHistory } from 'vue-router'

import { ROUTES } from '@/constants'

import GenerationView from '../views/GenerationView/GenerationView.vue'

const { HOME, TALK_TO_DATA } = ROUTES

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: HOME.path,
      name: HOME.name,
      component: GenerationView,
    },
    {
      path: TALK_TO_DATA.path,
      name: TALK_TO_DATA.name,
      component: () => import('../views/TalkToDataView.vue'),
    },
  ],
})

export default router
