<template lang="pug">
  v-list-tile(@click="$emit('userRequestClick', {id: id, userId: user_id})")
    v-list-tile-content
      v-list-tile-title {{ priorityString }} {{ userDisplayName }} {{ num }}個
      v-list-tile-sub-title 申請日時 {{ dateTimeString }}
</template>

<script>
import api from '../api'
import moment from 'moment'

export default {
  name: 'PlannerItemRequestUserDetail',
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
    },
    priority: {
      type: Number,
      default: -1
    }
  },
  computed: {
    dateTimeString: function () {
      return moment(this.updated_at).format('YYYY/MM/DD HH:mm:ss')
    },
    priorityString: function () {
      if (this.priority > 0) {
        return `(第${this.priority}希望)`
      }
      return ''
    }
  },
  asyncComputed: {
    userDisplayName: async function () {
      return api.getUserDisplayName(this.user_id)
    }
  }
}
</script>
