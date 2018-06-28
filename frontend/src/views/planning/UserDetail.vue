<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap)
      v-flex(d-block sm12 md4)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline {{ user.display_name }} (@{{ user.name }})
              v-card-text
                v-chip(v-if="user.permission === 1" color="green" text-color="white" small) プランナー
                v-chip(v-else-if="user.permission === 2" color="red" text-color="white" small) 管理人
                v-chip(v-if="user.entry_day1" color="primary" text-color="white" small) 1日目
                v-chip(v-if="user.entry_day2" color="primary" text-color="white" small) 2日目
                v-chip(v-if="user.entry_day3" color="primary" text-color="white" small) 3日目
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline リクエスト備考
              v-card-text
                request-note(v-for="noteId in notes" :key="noteId" :id="noteId")
      v-flex(d-flex sm12 md8)
        v-layout(row wrap)
          v-flex(d-flex xs12)
            v-card
              v-card-title.headline 希望順位
              v-container(fluid grid-list-xs)
                v-layout(row wrap)
                  v-flex(xs12 sm3 md3 lg3)
                    circle-priority-list(title="企業" :ids="priorities[0]")
                  v-flex(v-for="i in 3" xs12 sm3 md3 lg3 :key="i")
                    circle-priority-list(:title="i + '日目'" :ids="priorities[i]")
          v-flex(d-flex xs12)
            v-card
              v-card-title
                span.headline リクエストリスト
              v-container(fluid grid-list-xs)
                v-btn(block depressed color="primary" to="create-request" append) リクエスト追加
                v-radio-group(v-model="filter.day" row)
                  v-radio(:label="'企業' + '(' + requestedCircleCounts[0] + ')'" :value="0")
                  v-radio(v-for="i in 3" :label="i + '日目' + '(' + requestedCircleCounts[i] + ')'" :value="i" :key="i")
                v-layout(row wrap)
                  v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.circle_id")
                    planner-request-circle-item-list(v-bind="circle" :display-users="false" @itemClick="onItemClicked")
    v-dialog(v-model="editItemDialog.open" width=500 persistent)
      v-card
        v-card-title.headline {{ editItemDialog.itemName }}
        v-card-text
          v-form(v-model="editItemDialog.valid")
            v-text-field(v-model.number="editItemDialog.price" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="editItemDialog.sending")
            v-text-field(v-model.number="editItemDialog.num" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="editItemDialog.sending")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="editItemDialog.open = false" :disabled="editItemDialog.sending") キャンセル
          v-btn(flat color="red" @click.native="deleteRequest" :disabled="editItemDialog.sending" :loading="editItemDialog.deleting") 削除
          v-btn(flat color="primary" @click.native="editRequest" :disabled="editItemDialog.sending || !editItemDialog.valid" :loading="editItemDialog.editing") 修正

</template>

<script>
import api from '../../api'
import UserDisplayNameSpan from '../../components/UserDisplayNameSpan'
import CirclePriorityList from '../../components/CirclePriorityList'
import RequestNote from '../../components/RequestNote'
import PlannerRequestCircleItemList from '../../components/PlannerRequestCircleItemList'

export default {
  name: 'UserDetail',
  components: {
    UserDisplayNameSpan,
    CirclePriorityList,
    RequestNote,
    PlannerRequestCircleItemList
  },
  data: function () {
    return {
      id: null,
      user: {},
      prioritiesData: null,
      notes: [],
      requests: [],
      circleMap: new Map(),
      itemMap: new Map(),
      requestMap: new Map(),
      filter: {
        day: 0
      },
      editItemDialog: {
        circleId: null,
        itemId: null,
        requestId: null,
        open: false,
        sending: false,
        deleting: false,
        editing: false,
        itemName: '',
        price: -1,
        num: 1,
        valid: false
      },
      rules: {
        required: value => !!value || '必須項目です'
      }
    }
  },
  computed: {
    userId: function () {
      return parseInt(this.$route.params.id, 10)
    },
    filteredRequests: function () {
      return this.requests.filter(v => v.circle.day === this.filter.day)
    },
    requestedCircleCounts: function () {
      return this.requests.reduce((x, y) => {
        x[y.circle.day]++
        return x
      }, [0, 0, 0, 0])
    },
    priorities: function () {
      if (this.prioritiesData == null) {
        return [[], [], [], []]
      }
      return [this.prioritiesData.enterprise, this.prioritiesData.day1, this.prioritiesData.day2, this.prioritiesData.day3]
    }
  },
  watch: {
    '$route': 'reloadAll'
  },
  methods: {
    reloadAll: async function () {
      this.editItemDialog.open = false

      try {
        this.user = await api.getUser(this.userId)
        this.prioritiesData = await api.getUserPriority(this.userId)
        this.notes = await api.getUserRequestNotes(this.userId)
        const data = await api.getUserRequests(this.userId)
        this.circleMap.clear()
        this.itemMap.clear()
        this.requestMap.clear()
        for (const circle of data) {
          circle.circle = await api.getCircle(circle.circle_id)
          this.circleMap.set(circle.circle_id, circle)
          for (const item of circle.items) {
            this.itemMap.set(item.id, item)
            for (const request of item.requests) {
              this.requestMap.set(request.id, request)
            }
          }
        }
        this.requests = data
      } catch (e) {
        console.error(e)
      }
    },
    onItemClicked: function ({circleId, itemId}) {
      const item = this.itemMap.get(itemId)
      if (item) {
        const req = item.requests[0]
        this.editItemDialog.circleId = circleId
        this.editItemDialog.itemId = itemId
        this.editItemDialog.requestId = req.id
        this.editItemDialog.itemName = item.item.name
        this.editItemDialog.price = item.item.price === -1 ? '' : item.item.price
        this.editItemDialog.num = req.num

        this.editItemDialog.open = true
      }
    },
    editRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.editing = true
      try {
        const item = this.itemMap.get(this.editItemDialog.itemId)
        const req = item.requests[0]

        const p = this.editItemDialog.price === '' ? -1 : this.editItemDialog.price
        if (req.num !== this.editItemDialog.num) {
          await api.editRequest(this.editItemDialog.requestId, this.editItemDialog.num)
        }
        if (item.item.price !== p) {
          await api.patchCircleItemPrice(this.editItemDialog.itemId, p)
        }

        req.num = this.editItemDialog.num
        item.item.price = p
        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.editItemDialog.editing = false
      this.editItemDialog.sending = false
    },
    deleteRequest: async function () {
      this.editItemDialog.sending = true
      this.editItemDialog.deleting = true
      try {
        await api.deleteRequest(this.editItemDialog.requestId)

        this.requestMap.delete(this.editItemDialog.requestId)
        this.itemMap.delete(this.editItemDialog.itemId)
        const circle = this.circleMap.get(this.editItemDialog.circleId)
        if (circle) {
          let i = 0
          while (circle.items[i].id !== this.editItemDialog.itemId) i++
          circle.items.splice(i, 1)
          if (circle.items.length === 0) {
            let i = 0
            while (this.requests[i].circle_id !== circle.circle_id) i++
            this.requests.splice(i, 1)
            this.circleMap.delete(circle.circle_id)
          }
        }

        this.editItemDialog.open = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.editItemDialog.deleting = false
      this.editItemDialog.sending = false
    }
  },
  beforeRouteEnter: function (to, from, next) {
    next(async vm => {
      await vm.reloadAll()
    })
  }
}
</script>
