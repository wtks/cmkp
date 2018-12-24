import axios from 'axios'
import jwtDecode from 'jwt-decode'
import localforage from 'localforage'
import Vue from 'vue'
import VueApollo from 'vue-apollo'
import { createApolloClient } from 'vue-cli-plugin-apollo/graphql-client'
import store from './store'
import { onLogin, onLogout } from './vue-apollo'

Vue.use(VueApollo)
localforage.config({ name: 'cmkp' })

const getToken = () => localStorage.getItem('api_token')
const deleteToken = () => localStorage.removeItem('api_token')
const saveToken = (token) => localStorage.setItem('api_token', token)

let baseUrl = '/api/'
if (process.env.NODE_ENV === 'production') {
  baseUrl = process.env.VUE_APP_API_ENDPOINT + '/api/'
} else {
  baseUrl = 'http://localhost:3000/api/'
}

const { apolloClient, wsClient } = createApolloClient({
  httpEndpoint: baseUrl + 'graphql',
  persisting: false,
  websocketsOnly: false,
  ssr: false,
  getAuth: _ => `Bearer ${getToken()}`
})
apolloClient.wsClient = wsClient

let client
const initClient = (token) => {
  if (token) {
    const decoded = jwtDecode(token)
    store.commit('setUserId', parseInt(decoded.sub, 10))
    store.commit('setName', decoded.aud)
    client = axios.create({
      baseURL: baseUrl,
      headers: {
        'Authorization': 'Bearer ' + token
      }
    })
    onLogin(apolloClient)
  } else {
    client = axios.create({
      baseURL: baseUrl
    })
    onLogout(apolloClient)
  }
}

initClient(getToken())
export default {
  apolloProvider: () =>
    new VueApollo({
      defaultClient: apolloClient,
      defaultOptions: {
        $query: {
          fetchPolicy: 'cache-and-network'
        }
      },
      errorHandler (error) {
        // eslint-disable-next-line no-console
        console.log('%cError', 'background: red; color: white; padding: 2px 4px; border-radius: 3px; font-weight: bold;', error.message)
      }
    }),
  login: async (id, password) => {
    const res = await client.post('/login', {
      username: id,
      password: password
    })
    const token = res.data.token

    saveToken(token)
    initClient(token)
  },
  logout: () => {
    deleteToken()
    initClient(null)
  },
  getUserDisplayName: async (uid) => {
    const res = await client.get(`/users/${uid}`)
    return res.data.display_name
  },
  searchCircles: async (query, days = [0, 1, 2, 3]) => {
    try {
      const res = await client.get('/circles', { params: { q: query, days: days.join(',') } })
      return res.data
    } catch (e) {
      return []
    }
  },
  getCircleItems: async (cid) => {
    const res = await client.get(`/circles/${cid}/items`)
    return res.data
  },
  createCircleItem: async (cid, name, price) => {
    if (price == null || price === '') {
      price = -1
    }
    const res = await client.post('/items', {
      circle_id: cid,
      name: name,
      price: price
    })
    return res.data
  },
  patchCircleItemPrice: async (id, price) => {
    if (price == null || price === '') {
      price = -1
    }
    await client.patch(`/items/${id}/price`, {
      price: price
    })
  },
  createRequest: async (userId, itemId, num) => {
    const res = await client.post('/requests', {
      user_id: userId,
      item_id: itemId,
      num: num
    })
    return res.data
  },
  editRequest: async (rid, num) => {
    await client.patch(`/requests/${rid}`, {
      num: num
    })
  },
  deleteRequest: async (rid) => {
    await client.delete(`/requests/${rid}`)
  }
}
