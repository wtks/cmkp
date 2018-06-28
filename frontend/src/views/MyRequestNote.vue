<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text
        | リクエストに関して管理人に伝えたいことがあれば書いてください。
        v-form(ref="form" v-model="valid")
          v-textarea(v-model="newContent" label="備考" :rules="[rules.required]")
      v-card-actions
        v-spacer
        v-btn(depressed color="primary" @click="sendNote" :disabled="!valid || sending" :loading="sending") 送信

    v-layout(column)
      v-flex(xs12 v-for="noteId in notes" :key="noteId")
        v-card
          v-card-text
            request-note(:id="noteId")

</template>

<script>
import api from '../api'
import RequestNote from '../components/RequestNote'

export default {
  name: 'MyRequestNote',
  components: {
    RequestNote
  },
  data: function () {
    return {
      notes: [],
      valid: false,
      sending: false,
      newContent: '',
      rules: {
        required: value => !!value || '必須項目です'
      }
    }
  },
  mounted: async function () {
    await this.reloadMyNotes()
  },
  methods: {
    reloadMyNotes: async function () {
      try {
        this.notes = await api.getMyRequestNotes()
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    sendNote: async function () {
      this.sending = true
      try {
        const newId = await api.postMyRequestNotes(this.newContent)
        this.notes.unshift(newId)
        this.$refs.form.reset()
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.sending = false
    }
  }
}
</script>
