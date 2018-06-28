<template lang="pug">
  v-container(fluid grid-list-md)
    p
      user-display-name(:id="userId")
      | のリクエストを追加します。
    v-autocomplete(hide-no-data hide-selected :item-text="itemTextConverter" item-value="id" label="サークル選択" return-object placeholder="サークル名または作家名を入力" :items="searchResults" v-model="circle" :loading="searching" prepend-icon="search" :search-input.sync="searchQuery" clearable)
    v-layout(v-if="cid != null" row wrap)
      v-flex(sm5 md4)
        circle-detail-info(:id="cid")
      v-flex(sm7 md8)
        v-card
          v-card-title.headline 商品情報入力
          v-card-text
            v-form(v-model="valid")
              v-select(v-model="selectedItem" :items="items" label="商品名を選択" hint="希望の商品が無い場合は新規登録を選んでください" persistent-hint return-object single-line item-text="name" item-value="id")
              v-text-field(v-model="name" label="商品名を入力" hint="曖昧な商品名を入力しないでください。また複数の商品を一つにまとめて登録しないでください。(OK：新刊A, NG：新刊AとB)" required :rules="[rules.required]" maxLength="100" counter persistent-hint :disabled="selectedItem == null || isItemSelected")
              v-text-field(v-model.number="price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="selectedItem == null")
              v-text-field(v-model.number="num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="selectedItem == null")
          v-card-actions
            v-btn(block color="primary" :disabled="!valid || sending" :loading="sending" @click="createRequest") 登録
</template>

<script>
import CircleDetailInfo from '../../components/CircleDetailInfo'
import UserDisplayName from '../../components/label/UserDisplayName'
import { debounce } from 'throttle-debounce'
import api from '../../api'

export default {
  name: 'CreateUserRequest',
  components: {
    CircleDetailInfo,
    UserDisplayName
  },
  data: function () {
    return {
      searchResults: [],
      searchQuery: null,
      searching: false,
      circle: null,
      sending: false,
      selectedItem: null,
      name: '',
      price: '',
      num: 1,
      items: [],
      rules: {
        required: value => !!value || '必須項目です'
      },
      valid: false
    }
  },
  computed: {
    isItemSelected: function () {
      return this.selectedItem != null && this.selectedItem.id != null
    },
    cid: function () {
      return this.circle == null ? null : this.circle.id
    },
    userId: function () {
      return parseInt(this.$route.params.id, 10)
    }
  },
  methods: {
    createRequest: async function () {
      this.sending = true
      try {
        let itemId
        if (!this.isItemSelected) {
          itemId = (await api.createCircleItem(this.cid, this.name, this.price)).id
        } else {
          itemId = this.selectedItem.id
          if (this.price !== this.selectedItem.price) {
            await api.patchCircleItemPrice(this.selectedItem.id, this.price)
          }
        }
        await api.createRequest(this.userId, itemId, this.num)
        this.$router.push('/planning/users/' + this.userId)
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
    searchCircles: debounce(500, async function () {
      if (this.circle !== null) return
      this.searching = true
      try {
        this.searchResults = await api.searchCircles(this.searchQuery)
      } catch (e) {
        console.error(e)
      } finally {
        this.searching = false
      }
    }),
    itemTextConverter: function (val) {
      let str = ''
      if (val.day > 0) {
        str += `${val.day}日目 `
      } else {
        str += '企業 '
      }
      str += (val.day !== 0 ? val.hall + val.block + val.space : val.hall + val.space) + ' '
      str += val.name + ' '
      str += val.author
      return str
    },
    reload: async function () {
      this.selectedItem = null
      this.items = []
      if (this.cid != null) {
        this.sending = true
        try {
          const items = await api.getCircleItems(this.cid)
          this.items = [
            {
              name: '新規登録',
              id: null
            },
            ...items
          ]
        } catch (e) {
          console.error(e)
          if (e.response) {
            this.$bus.$emit('error', e.response.data.message)
          } else {
            this.$bus.$emit('error')
          }
        }
        this.sending = false
      }
    }
  },
  watch: {
    searchQuery: 'searchCircles',
    circle: 'reload',
    selectedItem: function () {
      if (this.isItemSelected) {
        this.name = this.selectedItem.name
        this.price = this.selectedItem.price >= 0 ? this.selectedItem.price : ''
        this.num = 1
      } else {
        this.name = ''
        this.price = ''
        this.num = 1
      }
    }
  }
}
</script>
