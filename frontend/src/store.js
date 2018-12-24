import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    userId: null,
    name: null,
    role: null,
    user: null
  },
  getters: {
    loggedIn: state => state.userId != null,
    myName: state => state.userId != null ? state.name : '',
    isAdmin: state => state.role != null && state.role === 'ADMIN',
    isPlanner: state => state.role != null && (state.role === 'ADMIN' || state.role === 'PLANNER')
  },
  mutations: {
    setUserId: function (state, id) {
      state.userId = id
    },
    setName: function (state, v) {
      state.name = v
    },
    setRole: function (state, role) {
      state.role = role
    }
  },
  actions: {}
})
