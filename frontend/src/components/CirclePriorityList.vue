<template lang="pug">
  v-card(flat)
    v-card-title {{ title }}
    v-list
      v-list-tile(v-for="i in 5" :key="i")
        v-list-tile-content
          v-list-tile-title
            | {{ i }}.&nbsp;
            router-link(v-if="circles[i-1] !== null" :to="`/circles/${circles[i-1].id}`" exact)
              span {{ circles[i-1].name }}
</template>

<script>

export default {
  name: 'CirclePriorityList',
  props: {
    title: {
      type: String,
      default: ''
    },
    circles: {
      type: Array,
      required: true
    }
  }
}
</script>
