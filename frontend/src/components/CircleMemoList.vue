<template lang="pug">
  v-card
    v-card-title.headline サークルメモ
    v-card-text
      div(v-for="(id, index) in memos" :key="id")
        circle-memo(:id="id" @deleted="onMemoDeleted")
        v-divider(v-if="index + 1 < memos.length")
    v-card-actions
      v-spacer
      v-dialog(v-model="dialog" persistent)
        v-btn(slot="activator" depressed color="primary") メモを書く
        v-card
          v-card-title.headline メモを作成
          v-card-text
            | メモは全員に公開されます。
            v-form(v-model="dialogValid")
              v-textarea(label="内容" v-model="newMemo" :rules="[v => !!v || '内容を入力してください']" required)
          v-card-actions
            v-spacer
            v-btn(flat @click.native="dialog = false; newMemo = ''") キャンセル
            v-btn(flat :disabled="!dialogValid || sending" :loading="sending" @click="createMemo") 作成

</template>

<script>
import api from '../api'
import CircleMemo from './CircleMemo'

export default {
  name: 'CircleMemoList',
  components: {
    CircleMemo
  },
  data: function () {
    return {
      loading: true,
      sending: false,
      dialog: false,
      dialogValid: false,
      newMemo: '',
      memos: []
    }
  },
  props: {
    id: {
      type: Number,
      required: true
    }
  },
  mounted: async function () {
    await this.reloadMemos()
  },
  watch: {
    id: async function () {
      await this.reloadMemos()
    }
  },
  methods: {
    reloadMemos: async function () {
      this.loading = true
      this.memos = await api.getCircleMemos(this.id)
      this.loading = false
    },
    createMemo: async function () {
      this.sending = true
      try {
        await api.createCircleMemo(this.id, this.newMemo)
        this.dialog = false
        this.newMemo = ''
        await this.reloadMemos()
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
    onMemoDeleted: function (id) {
      let i = 0
      while (this.memos[i] !== id) i++
      this.memos.splice(i, 1)
    }
  }
}
</script>
