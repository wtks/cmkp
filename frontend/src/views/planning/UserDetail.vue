<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap v-if="!$apollo.queries.fetchData.loading")
      v-flex(d-block sm12 md4)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline {{ user.displayName }} (@{{ user.name }})
              v-card-text
                v-chip(v-if="user.role === 'PLANNER'" color="green" text-color="white" small) プランナー
                v-chip(v-else-if="user.role === 'ADMIN'" color="red" text-color="white" small) 管理人
                v-chip(v-for="day in user.entryDays" :key="day" color="primary" text-color="white" small) {{ day }}日目
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline リクエスト備考
              v-card-text
                div(v-for="note in notes" :key="note.id")
                  span.caption {{ formatDatetime(note.updatedAt) }}
                  br
                  span.body-1(style="white-space: pre-wrap;word-wrap: break-word;" v-text="note.content" v-linkified)
      v-flex(d-flex sm12 md8)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline 希望順位
              v-container(fluid grid-list-xs)
                v-layout(row wrap)
                  v-flex(xs12 sm3 md3 lg3)
                    circle-priority-list(title="企業" :circles="priorities[0]")
                  v-flex(v-for="i in 3" xs12 sm3 md3 lg3 :key="i")
                    circle-priority-list(:title="`${i}日目`" :circles="priorities[i]")
          v-flex(d-flex xs12)
            v-card
              v-card-title
                span.headline リクエストリスト
              v-container(fluid grid-list-xs)
                v-btn(block depressed color="primary" to="create-request" append) リクエスト追加
                v-radio-group(v-model="filter.day" row)
                  v-radio(:label="`企業(${requestedCircleCounts[0]})`" :value="0")
                  v-radio(v-for="i in 3" :label="`${i}日目(${requestedCircleCounts[i]})`" :value="i" :key="i")
                v-layout(row wrap)
                  v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.id")
                    v-card
                      v-card-title.headline.lighten-4(:class="[{'orange': circle.locationType === 1}, {'red': circle.locationType === 2}, {'green': circle.locationType === 0}]")
                        router-link(:to="`/circles/${circle.id}`" style="text-decoration: none;") {{ circle.locationString }} {{ circle.name }}
                      v-divider
                      v-list
                        v-list-tile(v-for="item in circle.items" :key="item.id" @click="onItemClicked(circle, item)")
                          v-list-tile-content
                            v-list-tile-title {{ item.name }}
                            v-list-tile-sub-title {{ priceString(item.price) }} × {{ item.request.num }}個
    v-dialog(v-model="editItemDialog.open" width=500 persistent)
      v-card
        v-card-title.headline {{ editItemDialog.origItem.name }}
        v-card-text
          v-form(v-model="editItemDialog.valid")
            v-text-field(v-model.number="editItemDialog.price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="editItemDialog.sending")
            v-text-field(v-model.number="editItemDialog.num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="editItemDialog.sending")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="editItemDialog.open = false" :disabled="editItemDialog.sending") キャンセル
          v-btn(flat color="red" @click.native="deleteRequest" :disabled="editItemDialog.sending" :loading="editItemDialog.deleting") 削除
          v-btn(flat color="primary" @click.native="editRequest" :disabled="editItemDialog.sending || !editItemDialog.valid" :loading="editItemDialog.editing") 修正

</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'
import updateItemPrice from '../../gql/updateItemPrice.gql'
import changeRequestNum from '../../gql/changeRequestNum.gql'
import deleteRequest from '../../gql/deleteRequest.gql'
import CirclePriorityList from '../../components/CirclePriorityList'

const getData = gql`
  query ($uid: Int!) {
    user(id: $uid) {
      id
      name
      displayName
      role
      entryDays
    }
    requestNotes(userId: $uid) {
      id
      content
      updatedAt
    }
    userRequestedCircles(userId: $uid) {
      id
      name
      author
      day
      locationString
      locationType
      items: requestedItems(userId: $uid) {
        id
        name
        price
        request: userRequest(userId: $uid) {
          id
          num
        }
      }
    }
    priority0: circlePriority(userId: $uid, day: 0) {
      circles {
        id
        name
      }
    }
    priority1: circlePriority(userId: $uid, day: 1) {
      circles {
        id
        name
      }
    }
    priority2: circlePriority(userId: $uid, day: 2) {
      circles {
        id
        name
      }
    }
    priority3: circlePriority(userId: $uid, day: 3) {
      circles {
        id
        name
      }
    }
  }
`

export default {
  name: 'UserDetail',
  components: {
    CirclePriorityList
  },
  props: {
    userId: {
      type: Number,
      required: true
    }
  },
  data: function () {
    return {
      fetchData: {
        user: {},
        requestNotes: [],
        userRequestedCircles: [],
        priority0: {
          circles: []
        },
        priority1: {
          circles: []
        },
        priority2: {
          circles: []
        },
        priority3: {
          circles: []
        }
      },
      filter: {
        day: 0
      },
      editItemDialog: {
        origItem: {
          name: ''
        },
        origCircle: null,
        open: false,
        sending: false,
        deleting: false,
        editing: false,
        price: -1,
        num: 1,
        valid: false
      },
      rules: {
        required: value => !!value || '必須項目です'
      }
    }
  },
  computed: {
    user: function () {
      return this.fetchData.user
    },
    notes: function () {
      return this.fetchData.requestNotes
    },
    priorities: function () {
      return [this.fetchData.priority0.circles,
        this.fetchData.priority1.circles,
        this.fetchData.priority2.circles,
        this.fetchData.priority3.circles]
    },
    circles: function () {
      return this.fetchData.userRequestedCircles
    },
    filteredRequests: function () {
      return this.circles.filter(v => v.day === this.filter.day)
    },
    requestedCircleCounts: function () {
      return this.circles.reduce((x, y) => {
        x[y.day]++
        return x
      }, [0, 0, 0, 0])
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      variables: function () {
        return {
          uid: this.userId
        }
      },
      update: data => data
    }
  },
  methods: {
    priceString: function (p) {
      if (p >= 0) {
        return p + '円'
      } else {
        return '価格未定'
      }
    },
    formatDatetime: function (dt) {
      return dayjs(dt).fromNow()
    },
    onItemClicked: function (circle, item) {
      this.editItemDialog.origCircle = circle
      this.editItemDialog.origItem = item
      this.editItemDialog.price = item.price === -1 ? '' : item.price
      this.editItemDialog.num = item.request.num
      this.editItemDialog.open = true
    },
    editRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.editing = true
      try {
        const item = this.editItemDialog.origItem
        const req = item.request

        const p = this.editItemDialog.price === '' ? -1 : this.editItemDialog.price
        if (req.num !== this.editItemDialog.num) {
          await this.$apollo.mutate({
            mutation: changeRequestNum,
            variables: {
              id: req.id,
              num: this.editItemDialog.num
            }
          })
        }
        if (item.price !== p) {
          await this.$apollo.mutate({
            mutation: updateItemPrice,
            variables: {
              id: item.id,
              price: p
            }
          })
        }

        req.num = this.editItemDialog.num
        item.price = p
        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.editItemDialog.editing = false
      this.editItemDialog.sending = false
    },
    deleteRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.deleting = true
      try {
        await this.$apollo.mutate({
          mutation: deleteRequest,
          variables: {
            id: this.editItemDialog.origItem.request.id
          }
        })

        const circle = this.editItemDialog.origCircle
        let i = 0
        while (circle.items[i].id !== this.editItemDialog.origItem.id) i++
        circle.items.splice(i, 1)

        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.editItemDialog.deleting = false
      this.editItemDialog.sending = false
    }
  }
}
</script>
