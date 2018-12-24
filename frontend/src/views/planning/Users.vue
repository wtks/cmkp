<template lang="pug">
  v-container(fluid grid-list-md)
    v-list
      v-list-tile(v-for="user in users" :key="user.id" :to="user.id.toString()" append)
        v-list-tile-content
          v-list-tile-title {{ user.displayName }}
</template>

<script>
import gql from 'graphql-tag'

const getUsers = gql`
  query {
    users {
      id
      displayName
    }
  }
`

export default {
  name: 'Users',
  data: function () {
    return {
      users: []
    }
  },
  apollo: {
    users: {
      query: getUsers,
      fetchPolicy: 'cache-and-network',
      update: data => data.users
    }
  }
}
</script>
