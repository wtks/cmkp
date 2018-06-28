<template lang="pug">
  div 第{{priority}}希望：
    span(v-for="(user, idx) in users" :key="user.user_id")
      user-display-name-span(:id="user.user_id")
      template(v-if="idx !== users.length - 1") ,&nbsp;
</template>

<script>
import UserDisplayNameSpan from './UserDisplayNameSpan'

export default {
  name: 'PlannerRequestCirclePriority',
  components: {
    UserDisplayNameSpan
  },
  props: {
    priority: {
      type: Number,
      required: true
    },
    users: {
      type: Array,
      required: true
    }
  }
}
</script>
