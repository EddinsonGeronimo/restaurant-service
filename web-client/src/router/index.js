import Vue from 'vue'
import VueRouter from 'vue-router'
import SyncView from '../views/SyncView.vue'
import axios from 'axios'
import VueAxios from 'vue-axios'

Vue.use(VueRouter)
Vue.use(VueAxios, axios)

const routes = [
  {
    path: '/',
    name: 'SyncView',
    component: SyncView
  },
  {
    path: '/buyers',
    name: 'BuyersView',

    component: () => import(/* webpackChunkName: "buyersview" */ '../views/BuyersView.vue')
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
