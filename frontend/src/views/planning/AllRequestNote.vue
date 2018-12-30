<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(column)
      v-flex(xs12 v-for="note in notes" :key="note.id")
        v-card
          v-card-text
            span.caption
              router-link(:to="`/planning/users/${note.user.id}`") {{ note.user.displayName }}
              | &nbsp;- {{ datetimeString(note.updatedAt) }}
            br
            span.body-1.user-content-text(v-text="note.content" v-linkified)

</template>

<script>
import gql from 'graphql-tag'
import dayjs from 'dayjs'

const getRequestNotes = gql`
  query {
    requestNotes {
      id
      user {
        id
        displayName
      }
      content
      createdAt
      updatedAt
    }
  }
`

export default {
  name: 'AllRequestNote',
  data: function () {
    return {
      notes: []
    }
  },
  apollo: {
    notes: {
      query: getRequestNotes,
      fetchPolicy: 'cache-and-network',
      update: data => data.requestNotes
    }
  },
  methods: {
    datetimeString: function (dt) {
      return dayjs(dt).fromNow()
    }
  }
}
</script>
