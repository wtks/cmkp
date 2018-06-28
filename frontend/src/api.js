import axios from 'axios'
import Dexie from 'dexie'
import moment from 'moment'
import localforage from 'localforage'

const db = new Dexie('cmkp_db')
db.version(1).stores({
  circles: 'id'
})
db.version(2).stores({
  users: 'id, &name'
}).upgrade(tx => {})
db.version(3).stores({
  circle_memos: 'id'
}).upgrade(tx => {})
db.version(4).stores({
  requestNotes: 'id'
}).upgrade(tx => {})

localforage.config({
  name: 'cmkp'
})

const getToken = () => localStorage.getItem('api_token')
const deleteToken = () => localStorage.removeItem('api_token')
const saveToken = (token) => localStorage.setItem('api_token', token)

let client
const initClient = (token) => {
  if (token) {
    client = axios.create({
      baseURL: '/api/',
      headers: {
        'Authorization': 'Bearer ' + token
      }
    })
  } else {
    client = axios.create({
      baseURL: '/api/'
    })
  }
}

initClient(getToken())
export default {
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
  getDeadlines: async () => {
    const res = await client.get('/deadlines')
    return res.data
  },
  setDeadline: async (i, time) => {
    await client.put('/deadlines', {
      day: i,
      datetime: moment(time)
    })
  },
  getMe: () => client.get('/me'),
  getUsers: async () => {
    return (await client.get('/users')).data
  },
  getUser: async (uid) => {
    return (await client.get(`/users/${uid}`)).data
  },
  getUserDisplayName: async (uid) => {
    const cache = await db.users.get(uid)
    if (cache != null) {
      return cache.display_name
    }

    const res = await client.get(`/users/${uid}`)
    try {
      await db.users.put({
        id: res.data.id,
        name: res.data.name,
        display_name: res.data.display_name
      })
    } catch (e) {
      console.error(e)
    }
    return res.data.display_name
  },
  getUserPriority: async (uid) => {
    return (await client.get(`/users/${uid}/circle-priorities`)).data
  },
  getUserRequests: async (uid) => {
    const res = await client.get(`/users/${uid}/requests`)
    return res.data
  },
  getUserRequestNotes: async (uid) => {
    const res = await client.get(`/users/${uid}/request-notes`)
    return res.data
  },
  createUser: async (name, displayName, password) => {
    return (await client.post('/users', {
      username: name,
      display_name: displayName,
      password: password
    })).data
  },
  editUserEntries: async (uid, days) => {
    await client.patch(`/users/${uid}/entry`, {
      day1: days.indexOf(1) >= 0,
      day2: days.indexOf(2) >= 0,
      day3: days.indexOf(3) >= 0
    })
  },
  changeUserPassword: async (uid, password) => {
    await client.patch(`/users/${uid}/password`, {
      password: password
    })
  },
  changeUserPermission: async (uid, permission) => {
    await client.patch(`/users/${uid}/permission`, {
      permission: permission
    })
  },
  searchCircles: async (query, days = [0, 1, 2, 3]) => {
    try {
      const res = await client.get('/circles', {params: {q: query, days: days.join(',')}})
      return res.data
    } catch (e) {
      return []
    }
  },
  getRequestedCircles: async (day = null) => {
    const req = {}
    if (day != null) {
      req.day = day
    }
    return (await client.get('/circles/requested', {params: req})).data
  },
  getCircle: async (cid) => {
    // キャッシュファースト
    const cache = await db.circles.get(cid)
    if (cache != null) {
      cache.cached = true
      return cache
    }

    const res = await client.get(`/circles/${cid}`)
    try {
      await db.circles.put(res.data) // キャッシュ
    } catch (e) {
      console.error(e)
    }
    res.data.cached = false
    return res.data
  },
  getCircleItem: async (id) => {
    const res = await client.get(`/items/${id}`)
    return res.data
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
  getAllRequests: async () => {
    const res = await client.get('/requests')
    return res.data
  },
  getMyRequests: async () => {
    const res = await client.get('/me/requests')
    return res.data
  },
  createMyRequest: async (itemId, num) => {
    const res = await client.post('/me/requests', {
      item_id: itemId,
      num: num
    })
    return res.data
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
  },
  getMyPriority: async () => {
    return (await client.get('/me/circle-priorities')).data
  },
  setMyPriority: async (day, cids) => {
    await client.post('/me/circle-priorities', {
      day: day,
      priority: cids
    })
  },
  getMyRequestNotes: async () => {
    const res = await client.get('/me/request-notes')
    return res.data
  },
  postMyRequestNotes: async (content) => {
    const res = await client.post('/me/request-notes', {
      content: content
    })
    return res.data.id
  },
  getRequestNote: async (id) => {
    // キャッシュファースト
    const cache = await db.requestNotes.get(id)
    if (cache != null) {
      cache.cached = true
      return cache
    }

    const res = await client.get(`/request-notes/${id}`)
    try {
      await db.requestNotes.put(res.data) // キャッシュ
    } catch (e) {
      console.error(e)
    }
    res.data.cached = false
    return res.data
  },
  getAllRequestNotes: async () => {
    const res = await client.get('/request-notes')
    return res.data
  },
  deleteRequestNote: async (id) => {
    await client.delete(`/request-notes/${id}`)
  },
  getCircleMemo: async (id) => {
    // キャッシュファースト
    const cache = await db.circle_memos.get(id)
    if (cache != null) {
      cache.cached = true
      return cache
    }

    const res = await client.get(`/circle-memos/${id}`)
    try {
      await db.circle_memos.put(res.data) // キャッシュ
    } catch (e) {
      console.error(e)
    }
    res.data.cached = false
    return res.data
  },
  getCircleMemos: async (cid) => {
    const res = await client.get(`/circles/${cid}/memos`)
    return res.data
  },
  createCircleMemo: async (cid, content) => {
    const res = await client.post(`/circles/${cid}/memos`, {content: content})
    return res.data.id
  },
  deleteCircleMemo: async (id) => {
    await client.delete(`/circle-memos/${id}`)
  }
}
