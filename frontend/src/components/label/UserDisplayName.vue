<template lang="pug">
  span {{ userDisplayName }}
</template>

<script>
import api from '../../api'

export default {
  name: 'UserDisplayName',
  props: {
    id: {
      type: Number,
      required: true
    }
  },
  asyncComputed: {
    userDisplayName: async function () {
      return api.getUserDisplayName(this.id)
    }
  }
}
</script>
