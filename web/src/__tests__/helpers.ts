import { createRouter, createWebHistory } from 'vue-router'

export function createTestingRouter() {
  const component = { template: '<div>Overview</div>' }
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/', component },
      { path: '/notifications', component },
      { path: '/inbox', component },
      { path: '/templates', component },
      { path: '/settings', component },
    ],
  })
}
