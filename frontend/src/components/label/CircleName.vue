<template lang="pug">
  span {{ displayString }}
</template>

<script>
import api from '../../api'

export default {
  name: 'CircleName',
  props: {
    id: {
      type: Number,
      required: true
    },
    location: {
      type: Boolean,
      default: false
    },
    day: {
      type: Boolean,
      default: false
    }
  },
  data: function () {
    return {
      circle: null
    }
  },
  computed: {
    displayString: function () {
      if (this.circle == null) {
        return ''
      }
      let str = ''
      if (this.day) {
        if (this.circle.day > 0) {
          str += `${this.circle.day}日目 `
        } else {
          str += '企業 '
        }
      }
      if (this.location) {
        str += this.circle.day !== 0 ? this.circle.hall + this.circle.block + this.circle.space : this.circle.hall + this.circle.space
        str += ' '
      }
      str += this.circle.name
      return str
    }
  },
  created: async function () {
    await this.reload()
  },
  watch: {
    'id': 'reload'
  },
  methods: {
    reload: async function () {
      try {
        this.circle = await api.getCircle(this.id)
      } catch (e) {
        console.error(e)
      }
    }
  }
}
</script>
