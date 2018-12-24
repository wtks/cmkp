<template lang="pug">
  div
    span.caption {{ username }} - {{ datetimeString }}
    br
    span.body-1(style="white-space: pre-wrap;word-wrap: break-word;" v-text="content" v-linkified)
    template(v-if="isMine")
      v-flex(text-xs-right)
        v-dialog(v-model="deleteDialog" persistent width=500)
          v-btn(slot="activator" depressed small @click.stop="deleteDialog = true") 削除
          v-card
            v-card-title 本当に削除して良いですか？
            v-card-actions
              v-spacer
              v-btn(flat @click.native="deleteDialog = false") キャンセル
              v-btn(color="red" flat :disabled="sending" :loading="sending" @click.native="deleteMemo") OK

</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'

const deleteCircleMemo = gql`
  mutation ($id: Int!) {
    deleteCircleMemo(id: $id)
  }
`

export default {
  name: 'CircleMemo',
  data: function () {
    return {
      memo: null,
      deleteDialog: false,
      sending: false
    }
  },
  props: {
    id: {
      type: Number,
      required: true
    },
    user: {
      type: Object,
      required: true
    },
    userId: {
      type: Number,
      required: true
    },
    content: {
      type: String,
      required: true
    },
    createdAt: {
      type: String,
      required: true
    },
    updatedAt: {
      type: String,
      required: true
    }
  },
  computed: {
    datetimeString: function () {
      return dayjs(this.updatedAt).fromNow()
    },
    isMine: function () {
      return this.$store.state.userId === this.userId
    },
    username: function () {
      return this.user.displayName
    }
  },
  methods: {
    deleteMemo: async function () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: deleteCircleMemo,
          variables: {
            id: this.id
          }
        })
        this.$emit('deleted', this.id)
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
      this.deleteDialog = false
    }
  }
}
</script>
