<template lang="pug">
  div
    template(v-if="!error")
      span.caption {{ username }} - {{ datetimeString }}
      br
      span.body-1(style="white-space: pre-wrap;word-wrap: break-word;" v-text="content" v-linkified)
    template(v-else)
      span.caption 不明なユーザー
      br
      span.body-1(style="white-space: pre-wrap;word-wrap: break-word;") エラーが発生しました。

</template>

<script>
import api from '../api'
import moment from 'moment'

export default {
  name: 'RequestNote',
  data: function () {
    return {
      note: null,
      error: false
    }
  },
  props: {
    id: {
      type: Number,
      required: true
    }
  },
  computed: {
    datetimeString: function () {
      if (this.note == null) return ''
      return moment(this.note.created_at).fromNow()
    },
    content: function () {
      if (this.note == null) return '読み込み中'
      return this.note.content
    }
  },
  asyncComputed: {
    username: async function () {
      if (this.note == null) {
        return '読み込み中'
      }
      const name = await api.getUserDisplayName(this.note.user_id)
      return name
    }
  },
  created: async function () {
    await this.reloadNote()
  },
  watch: {
    id: 'reloadNote'
  },
  methods: {
    reloadNote: async function () {
      this.error = false
      try {
        this.note = await api.getRequestNote(this.id)
      } catch (e) {
        console.error(e)
        this.error = true
      }
    }
  }
}
</script>
