<template lang="pug">
  v-list-tile(@click="$emit('itemClick', id)")
    v-list-tile-content
      v-list-tile-title {{ item.name }}
      v-list-tile-sub-title {{ priceString }} × 計{{ requestedNum }}個
      v-list-tile-sub-title(v-if="displayUsers")
        span(v-for="(request, idx) in requests" :key="request.id")
          planner-item-request-user(v-bind="request")
          template(v-if="idx !== requests.length - 1") ,&nbsp;

</template>

<script>
import PlannerItemRequestUser from './PlannerItemRequestUser'

export default {
  name: 'PlannerRequestCircleItem',
  components: {
    PlannerItemRequestUser
  },
  props: {
    id: {
      type: Number,
      required: true
    },
    item: {
      type: Object,
      required: true
    },
    requests: {
      type: Array,
      required: true
    },
    displayUsers: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    priceString: function () {
      if (this.item.price >= 0) {
        return this.item.price + '円'
      } else {
        return '価格未定'
      }
    },
    requestedNum: function () {
      let num = 0
      for (const req of this.requests) num += req.num
      return num
    }
  }
}
</script>
