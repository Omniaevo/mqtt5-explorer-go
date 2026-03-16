import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/HomePage.vue'
import MessagesPage from '../pages/MessagesPage.vue'
import ChartsPage from '../pages/ChartsPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomePage
  },
  {
    path: '/messages',
    name: 'messages',
    component: MessagesPage
  },
  {
    path: '/charts',
    name: 'charts',
    component: ChartsPage
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsPage
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
