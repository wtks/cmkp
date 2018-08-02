<template lang="pug">
  v-card(flat)
    v-card-title {{ title }}
    v-list
      v-list-tile(v-for="i in 5" :key="i")
        v-list-tile-content
          v-list-tile-title
            | {{ i }}.&nbsp;
            router-link(v-if="ids.length > i-1" :to="'/circles/'+ids[i-1]" exact)
              circle-name(:id="ids[i-1]")
</template>

<script>
import CircleName from './label/CircleName'

export default {
  name: 'CirclePriorityList',
  components: {CircleName},
  props: {
    title: {
      type: String,
      default: ''
    },
    ids: {
      type: Array,
      required: true
    }
  }
}
</script>
