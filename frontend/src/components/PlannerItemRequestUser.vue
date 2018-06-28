<template lang="pug">
  span {{ userDisplayName }}({{ num }})
</template>

<script>
import api from '../api'

export default {
  name: 'PlannerItemRequestUser',
  props: {
    id: {
      type: Number,
      required: true
    },
    user_id: {
      type: Number,
      required: true
    },
    num: {
      type: Number,
      required: true
    },
    created_at: {
      type: String,
      required: true
    },
    updated_at: {
      type: String,
      required: true
    }
  },
  asyncComputed: {
    userDisplayName: async function () {
      return api.getUserDisplayName(this.user_id)
    }
  }
}
</script>
