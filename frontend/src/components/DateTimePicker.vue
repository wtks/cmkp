<template lang="pug">
  v-container(fluid)
    v-layout(row wrap)
      v-flex
        v-dialog(ref="date" v-model="dateDialog" :return-value.sync="date" persistent lazy full-width width="290px")
          v-text-field(slot="activator" v-model="date" label="日時" prepend-icon="event" readonly)
          v-date-picker(v-model="date" scrollable)
            v-spacer
            v-btn(flat color="primary" @click="dateDialog = false") キャンセル
            v-btn(flat color="primary" @click="$refs.date.save(date);updateValue()") OK

      v-flex
        v-dialog(ref="time" v-model="timeDialog" :return-value.sync="time" persistent lazy full-width width="290px")
          v-text-field(slot="activator" v-model="time" label="時刻" prepend-icon="access_time" readonly)
          v-time-picker(v-model="time" v-if="timeDialog")
            v-spacer
            v-btn(flat color="primary" @click="timeDialog = false") キャンセル
            v-btn(flat color="primary" @click="$refs.time.save(time);updateValue()") OK

</template>

<script>
import dayjs from 'dayjs'

export default {
  name: 'DateTimePicker',
  props: ['value'],
  data: function () {
    return {
      dateDialog: false,
      timeDialog: false,
      time: null,
      date: null
    }
  },
  watch: {
    value: function () {
      if (this.value == null) {
        this.date = ''
        this.time = ''
      } else {
        const v = dayjs(this.value)
        this.date = v.format('YYYY-MM-DD')
        this.time = v.format('HH:mm')
      }
    }
  },
  methods: {
    updateValue: function () {
      if (this.date != null && this.time != null) {
        this.$emit('input', `${this.date}T${this.time}:00+09:00`)
      }
    }
  }
}
</script>
