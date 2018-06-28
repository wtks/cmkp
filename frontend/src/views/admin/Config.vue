<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-flex
        v-card
          v-card-title.headline リクエスト締め切り設定
          v-card-text
            span 企業
            date-time-picker(v-model="deadlines.enterprise" @input="updateDeadline(0, $event)")
            span 1日目
            date-time-picker(v-model="deadlines.day1" @input="updateDeadline(1, $event)")
            span 2日目
            date-time-picker(v-model="deadlines.day2" @input="updateDeadline(2, $event)")
            span 3日目
            date-time-picker(v-model="deadlines.day3" @input="updateDeadline(3, $event)")

</template>

<script>
import api from '../../api'
import DateTimePicker from '../../components/DateTimePicker'

export default {
  name: 'Config',
  components: {
    DateTimePicker
  },
  data: function () {
    return {
      deadlines: {
        enterprise: null,
        day1: null,
        day2: null,
        day3: null
      }
    }
  },
  mounted: async function () {
    await this.reloadDeadlines()
  },
  methods: {
    reloadDeadlines: async function () {
      try {
        const data = await api.getDeadlines()
        this.deadlines.enterprise = data.enterprise
        this.deadlines.day1 = data.day1
        this.deadlines.day2 = data.day2
        this.deadlines.day3 = data.day3
        this.$store.commit('setDeadlines', data)
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    updateDeadline: async function (i, time) {
      try {
        await api.setDeadline(i, time)
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    }
  }
}
</script>
