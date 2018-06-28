<template lang="pug">
  v-container(fluid)
    v-layout(align-center justify-center)
      v-flex(xs12)
        p.caption ※入力した内容はパスワード以外後から変更できません。
        v-form(v-model="valid" lazy-validation)
          v-text-field(v-model="name" :rules="[rules.name]" :counter="20" label="ユーザーID" required)
          v-text-field(v-model="display_name" :rules="[rules.required]" label="表示名(全員に公開されます)" required)
          v-text-field(v-model="password" :rules="[rules.password]" label="パスワード" :append-icon="visiblePassword ? 'visibility_off' : 'visibility'" :type="visiblePassword ? 'text' : 'password'" @click:append="visiblePassword = !visiblePassword" required)
          v-btn(color="primary" :disabled="!valid" @click.stop="dialog = true" block) 登録
    v-dialog(v-model="dialog" persistent width=500)
      v-card
        v-card-title.headline 確認
        v-card-text
          p.subheading ユーザーID
          p {{ name }}
          p.subheading 表示名
          p {{ display_name }}
        v-card-actions
          v-spacer
            v-btn(flat :disabled="sending" @click.native="dialog = false") キャンセル
            v-btn(color="primary" flat :disabled="sending" :loading="sending" @click.native="submit") OK

</template>

<script>
import api from '../../api'

export default {
  name: 'UserCreate',
  data: function () {
    return {
      valid: false,
      name: '',
      display_name: '',
      password: '',
      visiblePassword: false,
      sending: false,
      dialog: false,
      rules: {
        required: value => !!value || '必須項目です',
        name: value => /^[0-9a-zA-Z_-]{1,20}$/.test(value) || 'idは半角英数文字と_と-のみ使えます',
        password: value => /^[a-zA-Z0-9!#$%&()*+,.:;=?@[\]^_{}-]+$/.test(value) || 'パスワードは半角英数文字と記号のみ使えます'
      }
    }
  },
  methods: {
    submit: async function () {
      this.sending = true
      try {
        await api.createUser(this.name, this.display_name, this.password)
        this.$router.push('/admin/users')
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.dialog = false
      this.sending = false
    }
  }
}
</script>
