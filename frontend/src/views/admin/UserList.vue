<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text 右下の+ボタンからメンバー登録ができます。
    v-layout(row wrap)
      v-radio-group(v-model="filterDay" row)
        v-radio(label="全員" :value="null")
        v-radio(label="1日目" :value="1")
        v-radio(label="2日目" :value="2")
        v-radio(label="3日目" :value="3")
    v-layout(row wrap)
      template(v-for="user in filteredUsers")
        v-flex(xs12 sm6 md4 lg3)
          v-card
            v-card-title
              span.headline {{ user.display_name }} (@{{ user.name }})
            v-card-text
              v-chip(v-if="user.permission === 1" color="green" text-color="white" small) プランナー
              v-chip(v-else-if="user.permission === 2" color="red" text-color="white" small) 管理人
              v-chip(v-if="user.entry_day1" color="primary" text-color="white" small) 1日目
              v-chip(v-if="user.entry_day2" color="primary" text-color="white" small) 2日目
              v-chip(v-if="user.entry_day3" color="primary" text-color="white" small) 3日目
            v-card-actions
              v-btn(depressed small @click.stop="openEditEntryDialog(user)") 参加日程修正
              v-btn(depressed small @click.stop="openChangePasswordDialog(user)") パスワード変更
              v-btn(depressed small @click.stop="openChangePermissionDialog(user)") 権限変更
    v-btn(fixed dark fab bottom right color="blue darken-2" to="/admin/users/create")
      v-icon add
    v-dialog(v-model="editEntryDialog" width=500 persistent)
      v-card
        v-card-title.headline 参加日程修正
        v-card-text
          p {{ editUser && editUser.display_name }}
          v-layout(row wrap)
            v-checkbox(v-model="editEntries" :value="1" label="1日目")
            v-checkbox(v-model="editEntries" :value="2" label="2日目")
            v-checkbox(v-model="editEntries" :value="3" label="3日目")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="editEntryDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending" :loading="sending" @click.native="editEntry") OK
    v-dialog(v-model="changePasswordDialog" width=500 persistent)
      v-card
        v-card-title.headline パスワード変更
        v-card-text
          p {{ editUser && editUser.display_name }}
          v-form(v-model="valid" ref="passwordForm")
            v-text-field(v-model="newPassword" :append-icon="visiblePassword ? 'visibility_off' : 'visibility'" :type="visiblePassword ? 'text' : 'password'" label="新しいパスワード" :rules="[rules.password]" @click:append="visiblePassword = !visiblePassword")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="changePasswordDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending || !valid" :loading="sending" @click.native="changePassword") OK
    v-dialog(v-model="changePermissionDialog" width=500 persistent)
      v-card
        v-card-title.headline 権限変更
        v-card-text
          p {{ editUser && editUser.display_name }}
          v-select(v-model="newPermission" :items="permissions")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="changePermissionDialog = false") キャンセル
          v-btn(color="primary" flat :disabled="sending" :loading="sending" @click.native="changePermission") OK

</template>

<script>
import api from '../../api'

export default {
  name: 'UserList',
  data: function () {
    return {
      users: [],
      filterDay: null,
      sending: false,
      editEntryDialog: false,
      changePasswordDialog: false,
      changePermissionDialog: false,
      visiblePassword: false,
      valid: false,
      editUser: null,
      editEntries: [],
      newPassword: '',
      newPermission: 0,
      rules: {
        password: value => /^[a-zA-Z0-9!#$%&()*+,.:;=?@[\]^_{}-]+$/.test(value) || 'パスワードは半角英数文字と記号のみ使えます'
      },
      permissions: [
        {text: 'メンバー', value: 0},
        {text: 'プランナー', value: 1},
        {text: '管理人', value: 2}
      ]
    }
  },
  computed: {
    filteredUsers: function () {
      return this.users.filter((v, i, a) => {
        switch (this.filterDay) {
          case 1:
            return v.entry_day1
          case 2:
            return v.entry_day2
          case 3:
            return v.entry_day3
          default:
            return true
        }
      })
    }
  },
  mounted: async function () {
    await this.updateUserList()
  },
  methods: {
    updateUserList: async function () {
      try {
        this.users = await api.getUsers()
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    editEntry: async function () {
      this.sending = true
      try {
        await api.editUserEntries(this.editUser.id, this.editEntries)
        this.editEntryDialog = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      await this.updateUserList()
      this.sending = false
    },
    changePassword: async function () {
      this.sending = true
      try {
        await api.changeUserPassword(this.editUser.id, this.newPassword)
        this.changePasswordDialog = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.sending = false
    },
    changePermission: async function () {
      this.sending = true
      try {
        await api.changeUserPermission(this.editUser.id, this.newPermission)
        this.changePermissionDialog = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.sending = false
    },
    openEditEntryDialog: function (user) {
      this.editUser = user
      const days = []
      if (user.entry_day1) {
        days.push(1)
      }
      if (user.entry_day2) {
        days.push(2)
      }
      if (user.entry_day3) {
        days.push(3)
      }
      this.editEntries = days
      this.editEntryDialog = true
    },
    openChangePasswordDialog: function (user) {
      this.editUser = user
      this.newPassword = ''
      this.valid = false
      this.$refs.passwordForm.reset()
      this.changePasswordDialog = true
    },
    openChangePermissionDialog: function (user) {
      this.editUser = user
      this.newPermission = user.permission
      this.changePermissionDialog = true
    }
  }
}
</script>
