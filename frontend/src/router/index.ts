import { createRouter, createWebHashHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/home' },
    { path: '/home', component: () => import('@/views/Home.vue') },
    { path: '/goods', component: () => import('@/views/Goods.vue') },
    { path: '/goods/details', component: () => import('@/views/Details.vue') },
    { path: '/error', component: () => import('@/views/Error.vue') },
    {
      path: '/manager',
      component: () => import('@/views/Manager.vue'),
      children: [
        { path: '', redirect: 'goodsmanage' },
        { path: 'goodsmanage', component: () => import('@/components/GoodsManage.vue') },
        { path: 'uploadgoods', component: () => import('@/components/UploadGoods.vue') },
        { path: 'discountsmanage', component: () => import('@/components/DiscountsManage.vue') },
        { path: 'usersmanage', component: () => import('@/components/UsersManage.vue') },
        { path: 'ordersmanage', component: () => import('@/components/OrdersManage.vue') },
        { path: 'about', component: () => import('@/views/About.vue') },
      ],
    },
    {
      path: '/shoppingCart',
      component: () => import('@/views/ShoppingCart.vue'),
      meta: { requireAuth: true },
    },
    {
      path: '/collect',
      component: () => import('@/views/Collect.vue'),
      meta: { requireAuth: true },
    },
    {
      path: '/order',
      component: () => import('@/views/Order.vue'),
      meta: { requireAuth: true },
    },
    {
      path: '/confirmOrder',
      component: () => import('@/views/ConfirmOrder.vue'),
      meta: { requireAuth: true },
    },
    {
      path: '/payment',
      component: () => import('@/views/Payment.vue'),
      meta: { requireAuth: true },
    },
    { path: '/:pathMatch(.*)*', redirect: '/error' },
  ],
})

router.beforeEach((to) => {
  if (to.meta.requireAuth) {
    const userStore = useUserStore()
    if (!userStore.user) {
      userStore.setShowLogin(true)
      return false
    }
  }
})

export default router
