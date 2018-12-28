import '@babel/polyfill'
import Vue from 'vue'
import linkify from 'vue-linkify'
import VueBus from 'vue-bus'
import './vuetify'
import App from './App.vue'
import router from './router'
import store from './store'
import './registerServiceWorker'
import dayjs from 'dayjs'
import 'dayjs/locale/ja'
import relativeTime from 'dayjs/plugin/relativeTime'
import api from './api'

dayjs.locale('ja')
dayjs.extend(relativeTime)

Vue.config.productionTip = false

Vue.use(VueBus)
Vue.directive('linkified', linkify)

new Vue({
  router,
  store,
  apolloProvider: api.apolloProvider(),
  render: h => h(App)
}).$mount('#app')
