<template lang="pug">
  v-card
    v-card-title.headline(:class="[{'orange': circle.location_type === 1}, {'red': circle.location_type === 2}, {'green': circle.location_type === 0}, 'lighten-4']")
      router-link(:to="'/circles/'+circle_id" style="text-decoration: none;") {{ locationString }} {{ circleName }}
    v-card-text.blue-grey.lighten-5(v-if="priorities.length > 0 && displayUsers")
      planner-request-circle-priority(v-for="p in priorities" :key="p.priority" v-bind="p")
    v-divider
    v-list(:three-line="displayUsers")
      planner-request-circle-item(v-for="item in items" :key="item.id" v-bind="item" @itemClick="onItemClicked" :display-users="displayUsers")

</template>

<script>
import PlannerRequestCircleItem from './PlannerRequestCircleItem'
import PlannerRequestCirclePriority from './PlannerRequestCirclePriority'

export default {
  name: 'PlannerRequestCircleItemList',
  components: {
    PlannerRequestCircleItem,
    PlannerRequestCirclePriority
  },
  props: {
    circle_id: {
      type: Number,
      required: true
    },
    circle: {
      type: Object,
      required: true
    },
    priorities: {
      type: Array,
      required: true
    },
    items: {
      type: Array,
      required: true
    },
    displayUsers: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    locationString: function () {
      return this.circle.day !== 0 ? this.circle.hall + this.circle.block + this.circle.space : this.circle.hall + this.circle.space
    },
    circleName: function () {
      return this.circle.name
    }
  },
  methods: {
    onItemClicked: function (id) {
      this.$emit('itemClick', {
        circleId: this.circle_id,
        itemId: id
      })
    }
  }
}
</script>
