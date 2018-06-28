<template lang="pug">
  v-app
    v-navigation-drawer(persistent v-model="drawer" enable-resize-watcher fixed app temporary)
      v-list(v-if="loggedIn")
        v-list-tile
          v-list-tile-action
            v-icon person
          v-list-tile-content
            v-list-tile-title {{ myDisplayName }}
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
      drawer: false
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
    }
  }
}
</script>
