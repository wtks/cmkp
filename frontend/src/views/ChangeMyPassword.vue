<template lang="pug">
  v-container(fluid)
    v-layout(align-center justify-center)
      v-flex(xs12)
        v-alert(v-model="successAlert" dismissible type="success") パスワードの変更に成功しました。
        v-form(v-model="valid" ref="form")
          v-text-field(v-model="oldPassword" :rules="[rules.password]" label="現在のパスワード" type="password" required)
          v-text-field(v-model="newPassword" :rules="[rules.password]" label="新しいパスワード" type="password" required)
          v-text-field(v-model="confirmPassword" :rules="[rules.confirmPassword]" label="新しいパスワード(確認)" type="password" required)
          v-btn(block color="primary" :disabled="!valid || loading" :loading="loading" @click="changePassword") 変更
</template>

<script>
import gql from 'graphql-tag'

export default {
  name: 'ChangeMyPassword',
  data: function () {
    return {
      valid: false,
      successAlert: false,
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
      loading: false,
      rules: {
        password: value => /^[a-zA-Z0-9!#$%&()*+,.:;=?@[\]^_{}-]+$/.test(value) || 'パスワードは半角英数文字と記号のみ使えます',
        confirmPassword: value => this.newPassword === value || '新しいパスワードを正しく入力してください'
      }
    }
  },
  methods: {
    changePassword: async function () {
      this.loading = true
      try {
        await this.$apollo.mutate({
          mutation: gql`
            mutation ($old: String!, $new: String!) {
              changePassword(oldPassword: $old, newPassword: $new)
            }
          `,
          variables: {
            old: this.oldPassword,
            new: this.newPassword
          }
        })
        this.$refs.form.reset()
        this.successAlert = true
      } catch (e) {
        this.$bus.$emit('error', e.graphQLErrors[0].message)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
