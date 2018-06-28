import '@babel/polyfill'
import Vue from 'vue'
import AsyncComputed from 'vue-async-computed'
import linkify from 'vue-linkify'
import VueBus from 'vue-bus'
import './plugins/vuetify'
import App from './App.vue'
import router from './router'
import store from './store'
import './registerServiceWorker'

Vue.config.productionTip = false

Vue.use(AsyncComputed)
Vue.use(VueBus)
Vue.directive('linkified', linkify)

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
