<template lang="pug">
  v-card
    v-card-title.headline(:class="[{'orange': circle.location_type === 1}, {'red': circle.location_type === 2}, {'green': circle.location_type === 0}, 'lighten-4']")
      router-link(:to="'/circles/'+circle_id" style="text-decoration: none;") {{ locationString }} {{ circleName }}
    v-list
      v-list-tile(v-for="request in items" :key="request.id" @click="$emit('itemClick', request.id)")
        v-list-tile-content
          v-list-tile-title {{ request.item.name }}
          v-list-tile-sub-title {{ request.item.price > -1 ? request.item.price + '円' : '価格未定' }} × {{ request.num }}個

    v-card-actions
      p 合計：{{ totalCosts }}
      v-spacer
      v-btn(depressed :to="'/my-requests/create/'+circle_id" :disabled="isDeadlineOver") 商品追加

</template>

<script>
export default {
  name: 'RequestCircleItemList',
  props: {
    circle_id: {
      type: Number,
      required: true
    },
    circle: {
      type: Object,
      required: true
    },
    items: {
      type: Array,
      required: true
    }
  },
  computed: {
    locationString: function () {
      if (this.circle == null) {
        return ''
      }
      return this.circle.day !== 0 ? this.circle.hall + this.circle.block + this.circle.space : this.circle.hall + this.circle.space
    },
    circleName: function () {
      if (this.circle == null) {
        return '不明なサークル'
      }
      return this.circle.name
    },
    isDeadlineOver: function () {
      switch (this.circle.day) {
        case 0:
          return this.$store.getters.isEnterpriseDeadlineOver
        case 1:
          return this.$store.getters.isDay1DeadlineOver
        case 2:
          return this.$store.getters.isDay2DeadlineOver
        case 3:
          return this.$store.getters.isDay3DeadlineOver
        default:
          return false
      }
    },
    totalCosts: function () {
      let unknown = false
      let costs = 0
      for (const req of this.items) {
        if (req.item.price === -1) {
          unknown = true
        } else {
          costs += req.item.price * req.num
        }
      }

      if (unknown) {
        return `${costs}円 + α`
      } else {
        return `${costs}円`
      }
    }
  }
}
</script>
