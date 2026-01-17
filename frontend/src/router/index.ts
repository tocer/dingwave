import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/messages',
      children: [
        {
          path: 'messages',
          component: () => import('@/views/MessagesView.vue'),
        },
        {
          path: 'contacts',
          component: () => import('@/views/ContactsView.vue'),
        },
      ],
    },
  ],
})

export default router
