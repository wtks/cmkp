<template lang="pug">
  v-dialog(width=500 v-model="dialog")
    v-card
      v-card-title.headline.grey.lighten-2(primary-title) エラー
      v-card-text {{ errorMsg }}
      v-divider
      v-card-actions
        v-spacer
        v-btn.primary(flat @click.native="dialog = false") OK

</template>

<script>
export default {
  name: 'ErrorDialog',
  data: function () {
    return {
      dialog: false,
      errorMsg: ''
    }
  },
  created: function () {
    this.$bus.$on('error', this.receiveError)
  },
  beforeDestroy: function () {
    this.$bus.$off('error', this.receiveError)
  },
  methods: {
    receiveError: function (msg) {
      if (msg) {
        this.errorMsg = msg
      } else {
        this.errorMsg = '何か問題が発生しているようです。'
      }
      this.dialog = true
    }
  }
}
</script>
