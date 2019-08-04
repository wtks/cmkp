<template lang="pug">
  v-container(fluid fill-height)
    v-layout(align-center justify-center)
      v-flex(xs12 sm8 md4)
        v-card
          v-card-text
            v-form(ref="form" v-model="valid")
              v-text-field(prepend-icon="person" :rules="[v => !!v || 'IDを入力してください']" label="ID" type="text" v-model="username" required)
              v-text-field(prepend-icon="lock" :rules="[v => !!v || 'パスワードを入力してください']" label="Password" type="password" v-model="password" required)
          v-card-actions
            v-spacer
            v-btn(color="primary" :loading="loading" :disabled="!valid || loading" @click="submit") ログイン

</template>

<script>
import api from '../api'

export default {
  name: 'Login',
  data: function () {
    return {
      username: '',
      password: '',
      valid: false,
      loading: false
    }
  },
  methods: {
    async submit () {
      if (this.$refs.form.validate()) {
        this.loading = true
        try {
          await api.login(this.username, this.password)
          this.$router.push('/')
        } catch (e) {
          console.error(e)
          if (e.response) {
            this.$bus.$emit('error', e.response.data.message)
          } else {
            this.$bus.$emit('error')
          }
        }
        this.loading = false
      }
    }
  }
}
</script>
