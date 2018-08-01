<template lang="pug">
  v-app
    v-navigation-drawer(persistent v-model="drawer" enable-resize-watcher fixed app temporary)
      v-list(v-if="loggedIn")
        v-list-tile
          v-list-tile-action
            v-icon person
          v-list-tile-content
            v-list-tile-title {{ myDisplayName }}
        v-list-tile(@click.native="passwordDialog.open = true")
          v-list-tile-action
            v-icon vpn_key
          v-list-tile-content
            v-list-tile-title パスワード変更
        v-list-tile(@click="logout")
          v-list-tile-action
            v-icon power_settings_new
          v-list-tile-content
            v-list-tile-title ログアウト
        v-divider
        v-list-tile(to="/")
          v-list-tile-action
            v-icon home
          v-list-tile-content
            v-list-tile-title ホーム
        v-list-tile(to="/circles")
          v-list-tile-action
            v-icon search
          v-list-tile-content
            v-list-tile-title サークル検索
        v-list-tile(to="/my-requests" exact)
          v-list-tile-action
            v-icon shopping_basket
          v-list-tile-content
            v-list-tile-title マイリクエスト
        v-list-tile(to="/my-requests/notes" exact)
          v-list-tile-action
            v-icon messages
          v-list-tile-content
            v-list-tile-title リクエスト備考
        template(v-if="isPlanner")
          v-divider
          v-list-tile(to="/planning/all-requests")
            v-list-tile-action
              v-icon library_books
            v-list-tile-content
              v-list-tile-title 全リクエストリスト
          v-list-tile(to="/planning/all-request-notes")
            v-list-tile-action
              v-icon library_books
            v-list-tile-content
              v-list-tile-title 全リクエスト備考
          v-list-tile(to="/planning/users")
            v-list-tile-action
              v-icon assignment_ind
            v-list-tile-content
              v-list-tile-title ユーザー別詳細
        template(v-if="isAdmin")
          v-divider
          v-list-tile(to="/admin/users")
            v-list-tile-action
              v-icon people
            v-list-tile-content
              v-list-tile-title メンバーリスト
          v-list-tile(to="/admin/config")
            v-list-tile-action
              v-icon settings
            v-list-tile-content
              v-list-tile-title 設定
    v-toolbar(app)
      v-toolbar-side-icon(@click.stop="drawer = !drawer")
      v-btn(v-show="backPagePath !== ''" icon :to="backPagePath" exact)
        v-icon arrow_back
      v-btn(v-show="backPage" icon @click="$router.go(-1)")
        v-icon arrow_back
      v-toolbar-title {{ pageName }}
    v-content
      v-fade-transition
        router-view
    v-footer(app)
      span cmkp &copy; wtks 2018
    error-dialog
    v-dialog(v-model="passwordDialog.open" persistent width=500)
      v-card
        v-card-title パスワード変更
        v-card-text
          v-form(v-model="passwordDialog.valid" lazy-validation)
            v-text-field(v-model="passwordDialog.oldPassword" :rules="[rules.password]" label="現在のパスワード" type="password" required)
            v-text-field(v-model="passwordDialog.newPassword" :rules="[rules.password]" label="新しいパスワード" type="password" required)
            v-text-field(v-model="passwordDialog.confirmPassword" :rules="[rules.confirmPassword]" label="新しいパスワード(確認)" type="password" required)
        v-card-actions
          v-spacer
          v-btn(@click="passwordDialog.open = false") キャンセル
          v-btn(color="primary" :disabled="!passwordDialog.valid || loading" :loading="loading" @click.native="changePassword") 変更
</template>

<script>
import api from './api'
import { mapGetters } from 'vuex'
import ErrorDialog from './components/ErrorDialog'

export default {
  name: 'App',
  components: {
    ErrorDialog
  },
  data () {
    return {
      drawer: false,
      passwordDialog: {
        open: false,
        valid: false,
        oldPassword: '',
        newPassword: '',
        confirmPassword: '',
        loading: false
      },
      rules: {
        password: value => /^[a-zA-Z0-9!#$%&()*+,.:;=?@[\]^_{}-]+$/.test(value) || 'パスワードは半角英数文字と記号のみ使えます',
        confirmPassword: value => this.passwordDialog.newPassword === value || '新しいパスワードを正しく入力してください'
      }
    }
  },
  computed: {
    pageName: function () {
      return this.$route.name
    },
    backPagePath: function () {
      const last = this.$route.matched.length - 1
      if (last >= 0 && this.$route.matched[last].meta) {
        if (this.$route.matched[last].meta.backPage) {
          return this.$route.matched[last].meta.backPage
        }
      }
      return ''
    },
    backPage: function () {
      const last = this.$route.matched.length - 1
      if (last >= 0 && this.$route.matched[last].meta) {
        if (this.$route.matched[last].meta.backButton) {
          return this.$route.matched[last].meta.backButton
        }
      }
      return false
    },
    ...mapGetters([
      'myDisplayName',
      'isAdmin',
      'isPlanner',
      'loggedIn'
    ])
  },
  methods: {
    logout: function () {
      api.logout()
      location.reload(true)
    },
    changePassword: async function () {
      this.passwordDialog.loading = true
      try {
        await api.changeMyPassword(this.passwordDialog.oldPassword, this.passwordDialog.newPassword)
        this.passwordDialog.oldPassword = ''
        this.passwordDialog.newPassword = ''
        this.passwordDialog.open = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      } finally {
        this.passwordDialog.loading = false
      }
    }
  }
}
</script>
