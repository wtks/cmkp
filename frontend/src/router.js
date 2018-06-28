import Vue from 'vue'
import Router from 'vue-router'
import api from './api'
import store from './store'
import Root from './Root'
import Home from './views/Home.vue'
import Login from './views/Login'
import MyRequests from './views/MyRequests'
import CreateRequest from './views/CreateRequest'
import SearchCircle from './views/SearchCircle'
import CircleInfo from './views/CircleInfo'
import UserList from './views/admin/UserList'
import UserCreate from './views/admin/UserCreate'
import Config from './views/admin/Config'
import AllRequestList from './views/planning/AllRequestList'
import MyRequestNote from './views/MyRequestNote'
import UserDetail from './views/planning/UserDetail'
import AllRequestNote from './views/planning/AllRequestNote'
import Users from './views/planning/Users'
import CreateUserRequest from './views/planning/CreateUserRequest'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      component: Root,
      beforeEnter: async function (to, from, next) {
        if (store.getters.loggedIn) {
          next()
        } else {
          try {
            const res = await api.getMe()
            if (res.status === 200) {
              store.commit('setMe', res.data)
              next()
            } else {
              next('/login')
            }
          } catch (e) {
            next('/login')
          }
        }
      },
      children: [
        {
          path: '',
          name: 'ホーム',
          component: Home,
          beforeEnter: function (to, from, next) {
            store.dispatch('fetchDeadlines')
            next()
          }
        },
        {
          path: 'circles',
          name: 'サークル検索',
          component: SearchCircle
        },
        {
          path: 'circles/:cid(\\d+)',
          name: 'サークル詳細',
          component: CircleInfo,
          meta: {
            backButton: true
          },
          beforeEnter: function (to, from, next) {
            store.dispatch('fetchDeadlines')
            next()
          }
        },
        {
          path: 'my-requests',
          component: Root,
          children: [
            {
              path: '',
              name: 'マイリクエスト',
              component: MyRequests
            },
            {
              path: 'create/:cid(\\d+)',
              name: 'リクエスト作成',
              component: CreateRequest,
              meta: {
                backButton: true
              }
            },
            {
              path: 'notes',
              name: 'リクエスト備考',
              component: MyRequestNote
            }
          ],
          beforeEnter: function (to, from, next) {
            store.dispatch('fetchDeadlines')
            next()
          }
        },
        {
          path: 'planning',
          component: Root,
          children: [
            {
              path: ''
            },
            {
              path: 'all-requests',
              name: '全リクエスト一覧',
              component: AllRequestList
            },
            {
              path: 'users',
              name: 'メンバー別詳細',
              component: Users
            },
            {
              path: 'users/:id(\\d+)',
              name: 'メンバー詳細',
              component: UserDetail,
              meta: {
                backButton: true
              }
            },
            {
              path: 'users/:id(\\d+)/create-request',
              name: 'メンバーリクエスト作成',
              component: CreateUserRequest,
              meta: {
                backButton: true
              }
            },
            {
              path: 'all-request-notes',
              name: '全リクエスト備考',
              component: AllRequestNote
            }
          ],
          beforeEnter: async function (to, from, next) {
            if (store.getters.isPlanner) {
              next()
            } else {
              next('/')
            }
          }
        },
        {
          path: 'admin',
          component: Root,
          children: [
            {
              path: 'users',
              name: 'メンバーリスト',
              component: UserList
            },
            {
              path: 'users/create',
              name: 'メンバー登録',
              component: UserCreate,
              meta: {
                backPage: '/admin/users'
              }
            },
            {
              path: 'config',
              name: '設定',
              component: Config
            }
          ],
          beforeEnter: async function (to, from, next) {
            if (store.getters.isAdmin) {
              next()
            } else {
              next('/')
            }
          }
        }
      ]
    },
    {
      path: '/login',
      name: 'ログイン',
      component: Login,
      beforeEnter: async function (to, from, next) {
        if (store.getters.loggedIn) {
          next('/')
        } else {
          try {
            const res = await api.getMe()
            if (res.status === 200) {
              store.commit('setMe', res.data)
              next('/')
            } else {
              next()
            }
          } catch (e) {
            next()
          }
        }
      }
    }
  ]
})
