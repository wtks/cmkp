<template lang="pug">
  v-card
    v-card-title.headline
      | {{ circle.circle.name }}
      br
      | {{ item.item.name }}
    v-card-text
      span {{ priceString }} × 計{{ requestedNum }}個
    v-list(two-line)
      planner-item-request-user-detail(v-for="req in requests" :key="req.id" v-bind="req")

</template>

<script>
import PlannerItemRequestUserDetail from './PlannerItemRequestUserDetail'

export default {
  name: 'PlannerRequestCircleItemDetail',
  components: {
    PlannerItemRequestUserDetail
  },
  props: {
    circle: {
      type: Object,
      required: true
    },
    item: {
      type: Object,
      required: true
    }
  },
  computed: {
    priceString: function () {
      if (this.item.item.price >= 0) {
        return this.item.item.price + '円'
      } else {
        return '価格未定'
      }
    },
    requestedNum: function () {
      let num = 0
      for (const req of this.item.requests) num += req.num
      return num
    },
    requests: function () {
      return this.item.requests.map(r => {
        for (const priority of this.circle.priorities) {
          for (const user of priority.users) {
            if (r.user_id === user.user_id) {
              return {
                ...r,
                priority: priority.priority
              }
            }
          }
        }
        return r
      })
    }
  }
}
</script>
