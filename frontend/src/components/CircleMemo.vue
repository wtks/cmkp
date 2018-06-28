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
import api from '../api'
import moment from 'moment'

export default {
  name: 'CircleMemo',
  data: function () {
    return {
      memo: null,
      deleteDialog: false,
      sending: false
    }
  },
  asyncComputed: {
    username: async function () {
      if (this.memo == null) {
        return '読み込み中'
      }
      return (await api.getUser(this.memo.user_id)).display_name
    }
  },
  computed: {
    datetimeString: function () {
      if (this.memo == null) return ''
      return moment(this.memo.created_at).fromNow()
    },
    isMine: function () {
      if (this.memo == null) return false
      return this.$store.state.user.id === this.memo.user_id
    },
    content: function () {
      if (this.memo == null) return '読み込み中'
      return this.memo.content
    }
  },
  props: {
    id: {
      type: Number,
      required: true
    }
  },
  created: async function () {
    await this.reloadMemo()
  },
  watch: {
    id: async function () {
      await this.reloadMemo()
    }
  },
  methods: {
    reloadMemo: async function () {
      try {
        this.memo = await api.getCircleMemo(this.id)
      } catch (e) {
        console.error(e)
      }
    },
    deleteMemo: async function () {
      this.sending = true
      try {
        await api.deleteCircleMemo(this.id)
        this.$emit('deleted', this.id)
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.sending = false
      this.deleteDialog = false
    }
  }
}
</script>
