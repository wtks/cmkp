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

let baseUrl
if (process.env.NODE_ENV === 'production') {
  baseUrl = process.env.VUE_APP_API_ENDPOINT + '/api/'
} else {
  baseUrl = process.env.VUE_APP_API_ENDPOINT + '/api/' || 'http://localhost:3000/api/'
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
  }
}
