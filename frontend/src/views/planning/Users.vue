<template lang="pug">
  v-container(fluid grid-list-md)
    v-list
      v-list-tile(v-for="user in users" :key="user.id" :to="user.id.toString()" append)
        v-list-tile-content
          v-list-tile-title {{ user.display_name }}
</template>

<script>
import api from '../../api'

export default {
  name: 'Users',
  data: function () {
    return {
      users: []
    }
  },
  mounted: async function () {
    await this.reload()
  },
  methods: {
    reload: async function () {
      try {
        this.users = await api.getUsers()
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    }
  }
}
</script>
