<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text
        | リクエストに関して管理人に伝えたいことがあれば書いてください。
        v-form(ref="form" v-model="valid")
          v-textarea(v-model="newContent" label="備考" :rules="[v => !!v || '必須項目です']")
      v-card-actions
        v-spacer
        v-btn(depressed color="primary" @click="sendNote" :disabled="!valid || sending" :loading="sending") 送信

    v-layout(column)
      v-flex(xs12 v-if="$apollo.loading") Loading...
      v-flex(xs12 v-else v-for="note in notes" :key="note.id")
        v-card
          v-card-text
            span.caption {{ datetimeString(note.updatedAt) }}
            br
            span.body-1.user-content-text(v-text="note.content" v-linkified)
</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'

const getMyRequestNotes = gql`
  query {
    myRequestNotes {
      id
      content
      createdAt
      updatedAt
    }
  }
`

const sendNote = gql`
  mutation ($content: String!) {
    postRequestNote(content: $content) {
      id
      content
      createdAt
      updatedAt
    }
  }
`

export default {
  name: 'MyRequestNote',
  data: function () {
    return {
      notes: [],
      valid: false,
      sending: false,
      newContent: ''
    }
  },
  apollo: {
    notes: {
      query: getMyRequestNotes,
      fetchPolicy: 'cache-and-network',
      update: data => data.myRequestNotes
    }
  },
  methods: {
    datetimeString: function (dt) {
      return dayjs(dt).fromNow()
    },
    sendNote: async function () {
      this.sending = true
      try {
        await this.$apollo.mutate({
          mutation: sendNote,
          variables: {
            content: this.newContent
          },
          update: (store, { data: { postRequestNote } }) => {
            const data = store.readQuery({ query: getMyRequestNotes })
            data.myRequestNotes.unshift(postRequestNote)
            store.writeQuery({ query: getMyRequestNotes, data })
          }
        })
        this.$refs.form.reset()
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    }
  }
}
</script>
