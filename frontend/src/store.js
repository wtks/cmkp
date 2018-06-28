import Vue from 'vue'
import Vuex from 'vuex'
import moment from 'moment'
import api from './api'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    user: null,
    deadlines: null
  },
  getters: {
    loggedIn: state => state.user != null,
    myDisplayName: state => state.user != null ? state.user.display_name : '',
    isAdmin: state => state.user != null && state.user.permission >= 2,
    isPlanner: state => state.user != null && state.user.permission >= 1,
    isEnterpriseDeadlineOver: state => {
      if (state.deadlines == null) {
        return false
      }
      if (state.deadlines.enterprise == null) {
        return false
      }
      return moment(state.deadlines.enterprise).isBefore()
    },
    isDay1DeadlineOver: state => {
      if (state.deadlines == null) {
        return false
      }
      if (state.deadlines.day1 == null) {
        return false
      }
      return moment(state.deadlines.day1).isBefore()
    },
    isDay2DeadlineOver: state => {
      if (state.deadlines == null) {
        return false
      }
      if (state.deadlines.day2 == null) {
        return false
      }
      return moment(state.deadlines.day2).isBefore()
    },
    isDay3DeadlineOver: state => {
      if (state.deadlines == null) {
        return false
      }
      if (state.deadlines.day3 == null) {
        return false
      }
      return moment(state.deadlines.day3).isBefore()
    },
    enterpriseDeadline: state => {
      if (state.deadlines == null) {
        return null
      }
      if (state.deadlines.enterprise == null) {
        return null
      }
      return moment(state.deadlines.enterprise)
    },
    day1Deadline: state => {
      if (state.deadlines == null) {
        return null
      }
      if (state.deadlines.day1 == null) {
        return null
      }
      return moment(state.deadlines.day1)
    },
    day2Deadline: state => {
      if (state.deadlines == null) {
        return null
      }
      if (state.deadlines.day2 == null) {
        return null
      }
      return moment(state.deadlines.day2)
    },
    day3Deadline: state => {
      if (state.deadlines == null) {
        return null
      }
      if (state.deadlines.day3 == null) {
        return null
      }
      return moment(state.deadlines.day3)
    }
  },
  mutations: {
    setMe: function (state, user) {
      state.user = user
    },
    setDeadlines: function (state, payload) {
      state.deadlines = payload
    }
  },
  actions: {
    fetchDeadlines: async function ({commit}) {
      try {
        const res = await api.getDeadlines()
        commit('setDeadlines', res)
      } catch (e) {
        console.error(e)
      }
    }
  }
})
