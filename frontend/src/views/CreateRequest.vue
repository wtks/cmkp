<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap v-if="!$apollo.queries.fetchData.loading")
      v-flex(sm5 md4)
        circle-detail-info(v-bind="fetchData.circle")
      v-flex(sm7 md8)
        v-card
          v-card-title.headline 商品情報入力
          v-card-text
            v-form(v-model="valid")
              v-select(v-model="selectedItem" :items="selectionItems" label="商品名を選択" hint="希望の商品が無い場合は新規登録を選んでください" persistent-hint return-object single-line item-text="name" item-value="id")
              v-text-field(v-model="name" label="商品名を入力" hint="曖昧な商品名を入力しないでください。また複数の商品を一つにまとめて登録しないでください。" required :rules="[rules.required]" maxLength="100" counter persistent-hint :disabled="selectedItem == null || isItemSelected")
              v-text-field(v-model.number="price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="selectedItem == null")
              v-text-field(v-model.number="num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="selectedItem == null")
          v-card-actions
            v-btn(block color="primary" :disabled="!valid || sending" :loading="sending" @click="createRequest") 登録
</template>

<script>
import gql from 'graphql-tag'
import createRequest from '../gql/createRequest.gql'
import updateItemPrice from '../gql/updateItemPrice.gql'
import CircleDetailInfo from '../components/CircleDetailInfo'

const getData = gql`
  query ($cid: Int!) {
    circle(id: $cid) {
      id
      name
      author
      hall
      day
      block
      space
      locationType
      genre
      pixivId
      twitterId
      niconicoId
      website
    }
    items(circleId: $cid) {
      id
      name
      price
    }
  }
`

const createItem = gql`
  mutation ($cid: Int!, $name: String!, $price: Int!) {
    createItem(circleId: $cid, name: $name, price: $price) {
      id
    }
  }
`

export default {
  name: 'CreateRequest',
  components: {
    CircleDetailInfo
  },
  props: {
    cid: {
      type: Number,
      required: true
    }
  },
  data: function () {
    return {
      fetchData: {
        circle: null,
        items: []
      },
      sending: false,
      selectedItem: null,
      name: '',
      price: '',
      num: 1,
      rules: {
        required: value => !!value || '必須項目です'
      },
      valid: false
    }
  },
  apollo: {
    fetchData: {
      query: getData,
      fetchPolicy: 'cache-and-network',
      variables: function () {
        return {
          cid: this.cid
        }
      },
      update: data => data
    }
  },
  computed: {
    isItemSelected: function () {
      return this.selectedItem != null && this.selectedItem.id != null
    },
    selectionItems: function () {
      return [
        {
          name: '新規登録',
          id: null
        },
        ...this.fetchData.items
      ]
    }
  },
  methods: {
    createRequest: async function () {
      this.sending = true
      try {
        let itemId
        if (!this.isItemSelected) {
          itemId = (await this.$apollo.mutate({
            mutation: createItem,
            variables: {
              cid: this.cid,
              name: this.name,
              price: this.price !== '' ? this.price : -1
            }
          })).data.createItem.id
        } else {
          itemId = this.selectedItem.id
          if (this.price !== this.selectedItem.price) {
            await this.$apollo.mutate({
              mutation: updateItemPrice,
              variables: {
                id: itemId,
                price: this.price !== '' ? this.price : -1
              }
            })
          }
        }

        await this.$apollo.mutate({
          mutation: createRequest,
          variables: {
            userId: null,
            itemId: itemId,
            num: this.num
          }
        })
        this.$router.push(`/my-requests?day=${this.fetchData.circle.day}`)
      } catch (e) {
        console.error(e)
        this.$bus.$emit('error')
      }
      this.sending = false
    }
  },
  watch: {
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
