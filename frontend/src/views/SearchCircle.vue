<template lang="pug">
  v-container(fluid v-scroll="onScroll")
    v-text-field(:disabled="loading" :loading="loading" append-icon="search" v-model="query" placeholder="サークル名または作家名を入力" @click:append="getCircles" @keypress.enter="getCircles")
    v-layout(row wrap)
      v-checkbox(label="1日目" v-model="filters.day1")
      v-checkbox(label="2日目" v-model="filters.day2")
      v-checkbox(label="3日目" v-model="filters.day3")
      v-checkbox(label="企業" v-model="filters.enterprise")
    v-container(fluid grid-list-md)
      v-data-iterator(:items="circles" :loading="loading" content-tag="v-layout" row wrap hide-actions)
        v-flex(slot="item" slot-scope="props" xs12 sm6 md4 lg3)
          v-card(:to="'/circles/'+props.item.id" :class="[{'blue': props.item.day === 1}, {'teal': props.item.day === 2}, {'lime': props.item.day === 3}, {'orange': props.item.day === 0}, 'lighten-5']")
            v-card-text
              div.caption {{ getCircleLoc(props.item) }}
              div {{ props.item.name }} - {{ props.item.author }}
        p(slot="no-data") 見つかりませんでした
    v-fade-transition
      v-btn(fixed dark fab bottom right color="blue darken-2" @click="$vuetify.goTo(0)" v-show="offsetTop > 100")
        v-icon arrow_upward

</template>

<script>
import api from '../api'

export default {
  name: 'SearchCircle',
  data: function () {
    return {
      offsetTop: 0,
      loading: false,
      query: '',
      circles: [],
      filters: {
        day1: true,
        day2: true,
        day3: true,
        enterprise: true
      }
    }
  },
  watch: {
    'filters.day1': function () {
      this.getCircles()
    },
    'filters.day2': function () {
      this.getCircles()
    },
    'filters.day3': function () {
      this.getCircles()
    },
    'filters.enterprise': function () {
      this.getCircles()
    }
  },
  computed: {
    days: function () {
      const res = []
      if (this.filters.enterprise) {
        res.push(0)
      }
      if (this.filters.day1) {
        res.push(1)
      }
      if (this.filters.day2) {
        res.push(2)
      }
      if (this.filters.day3) {
        res.push(3)
      }
      return res
    }
  },
  methods: {
    getCircles: async function () {
      this.loading = true
      this.circles = await api.searchCircles(this.query, this.days)
      this.loading = false
    },
    getCircleLoc: function (circle) {
      if (circle.day === 0) {
        return '企業 ' + circle.hall + circle.space
      } else {
        return circle.day + '日目 ' + circle.hall + circle.block + circle.space
      }
    },
    onScroll (e) {
      this.offsetTop = window.pageYOffset || document.documentElement.scrollTop
    }
  }
}
</script>
