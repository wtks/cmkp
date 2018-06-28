<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-flex(sm5 md4)
        circle-detail-info(v-if="cid != null" :id="cid")
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
import CircleDetailInfo from '../components/CircleDetailInfo'
import api from '../api'

export default {
  name: 'CreateRequest',
  components: {
    CircleDetailInfo
  },
  data: function () {
    return {
      cid: null,
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
        await api.createMyRequest(itemId, this.num)
        this.$router.push('/my-requests?day=' + this.circle.day)
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
    reloadCircleItems: async function () {
      this.sending = true
      this.selectedItem = null
      try {
        this.circle = await api.getCircle(this.cid)
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
    },
    fetchData: async function () {
      this.cid = parseInt(this.$route.params.cid, 10)
      await this.reloadCircleItems()
    }
  },
  watch: {
    '$route': 'fetchData',
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
  },
  created: function () {
    this.fetchData()
  }
}
</script>
