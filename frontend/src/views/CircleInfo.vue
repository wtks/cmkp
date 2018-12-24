<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-flex(v-if="$apollo.queries.fetchData.loading") Loading...
      template(v-else)
        v-flex
          circle-detail-info(v-bind="fetchData.circle")
        v-flex
          v-card
            v-card-title.headline サークルメモ
            v-card-text
              div(v-for="(memo, index) in fetchData.circleMemos" :key="memo.id")
                circle-memo(v-bind="memo" @deleted="onMemoDeleted")
                v-divider(v-if="index + 1 < fetchData.circleMemos.length")
            v-card-actions
              v-spacer
              v-dialog(v-model="dialog" persistent)
                v-btn(slot="activator" depressed color="primary") メモを書く
                v-card
                  v-card-title.headline メモを作成
                  v-card-text
                    | メモは全員に公開されます。
                    v-form(v-model="dialogValid")
                      v-textarea(label="内容" v-model="newMemo" :rules="[v => !!v || '内容を入力してください']" required)
                  v-card-actions
                    v-spacer
                    v-btn(flat @click.stop="dialog = false; newMemo = ''") キャンセル
                    v-btn(flat :disabled="!dialogValid || sending" :loading="sending" @click="createMemo") 作成
    v-btn(block color="success" :disabled="isDeadlineOver" :to="`/my-requests/create/${cid}`") リクエストを作成 {{ isDeadlineOver ? '(締め切りました)' : '' }}
</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'
import CircleDetailInfo from '../components/CircleDetailInfo'
import CircleMemo from '../components/CircleMemo'

const getData = gql`
  query ($cid: Int!) {
    circle(id: $cid) {
      id
      name
      author
      hall
      day
      block
      space
      locationType
      genre
      pixivId
      twitterId
      niconicoId
      website
    }
    circleMemos(circleId: $cid) {
      id
      userId
      user {
        displayName
      }
      content
      createdAt
      updatedAt
    },
    day0: deadline(day: 0)
    day1: deadline(day: 1)
    day2: deadline(day: 2)
    day3: deadline(day: 3)
  }
`

const createCircleMemo = gql`
  mutation ($cid: Int!, $content: String!) {
    postCircleMemo(circleId: $cid, content: $content) {
      id
      userId
      user {
        displayName
      }
      content
      createdAt
      updatedAt
    }
  }
`

export default {
  name: 'CircleInfo',
  components: {
    CircleDetailInfo,
    CircleMemo
  },
  props: {
    cid: {
      type: Number,
      required: true
    }
  },
  data: function () {
    return {
      fetchData: {
        circle: {
          day: null
        },
        circleMemos: [],
        day0: null,
        day1: null,
        day2: null,
        day3: null
      },
      sending: false,
      dialog: false,
      dialogValid: false,
      newMemo: ''
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      variables: function () {
        return {
          cid: this.cid
        }
      },
      update: data => data
    }
  },
  computed: {
    isDeadlineOver () {
      switch (this.fetchData.circle.day) {
        case 0:
          return dayjs(this.fetchData.day0).isBefore(dayjs())
        case 1:
          return dayjs(this.fetchData.day1).isBefore(dayjs())
        case 2:
          return dayjs(this.fetchData.day2).isBefore(dayjs())
        case 3:
          return dayjs(this.fetchData.day3).isBefore(dayjs())
        default:
          return false
      }
    }
  },
  methods: {
    async createMemo () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: createCircleMemo,
          variables: {
            cid: this.cid,
            content: this.newMemo
          },
          update: (store, { data: { postCircleMemo } }) => {
            const data = store.readQuery({ query: getData, variables: { cid: this.cid } })
            data.circleMemos.unshift(postCircleMemo)
            store.writeQuery({ query: getData, variables: { cid: this.cid }, data })
          }
        })
        this.dialog = false
        this.newMemo = ''
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    },
    onMemoDeleted (id) {
      let i = 0
      while (this.fetchData.circleMemos[i].id !== id) i++
      this.fetchData.circleMemos.splice(i, 1)
    }
  }
}
</script>
