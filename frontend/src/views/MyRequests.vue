<template lang="pug">
  v-container(fluid grid-list-md)
    v-card
      v-card-text
        | 右下の+ボタンまたはサークル別の商品追加ボタンからリクエストを作成できます。
        br
        | 登録商品を押すと、登録情報を修正・削除できます。サークル名を押すとサークル詳細を確認できます。
        v-select(label="日程" :items="daySelectItems" v-model="selectedDay")
        p 合計：{{ circleCount }}サークル, {{ totalCosts }}
        v-btn(color="primary" depressed block :disabled="isSelectedDayDeadlineOver" @click="openEditPriorityDialog") 希望優先順位を設定する

    v-layout(row wrap)
      v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.circle_id")
        request-circle-item-list(v-bind="circle" @itemClick="openEditDialog")

    v-btn(fixed dark fab bottom right color="blue darken-2" to="/circles")
      v-icon add

    v-dialog(v-model="dialog" width=500 persistent)
      v-card
        v-card-title.headline {{ editTarget != null ? editTarget.item.name : ''}}
        v-card-text
          v-form(v-model="valid")
            v-text-field(v-model.number="editPrice" label="単体価格" hint="決定していない場合は空欄にしてください" type="number" min="0" max="50000" persistent-hint :disabled="sending")
            v-text-field(v-model.number="editNum" label="個数" type="number" min="1" max="99" required :rules="[rules.required]" :disabled="!numEditable || sending")
        v-card-actions
          v-spacer
          v-btn(flat @click.native="dialog = false" :disabled="sending") キャンセル
          v-btn(flat color="red" @click.native="deleteRequest" :disabled="sending || !numEditable" :loading="deleting") 削除
          v-btn(flat color="primary" @click.native="editRequest" :disabled="sending || !valid" :loading="editing") 修正

    v-dialog(v-model="priorityDialog" width=500 persistent)
      v-card
        v-card-title.headline {{ selectedDay }}の希望優先順位
        v-card-text 上位5つまで設定できます。
        v-list
          v-list-tile(v-for="(p, i) of editPriorities" :key="p.id")
            v-list-tile-content
              v-list-tile-title 第{{ i + 1 }}希望 {{ p.name }}
            v-list-tile-action
              v-btn(icon @click="removePriorityCircle(i)")
                v-icon clear
        v-card-text
          v-select(v-model="editPrioritySelectCircle" :items="prioritySelectItems" placeholder="ここから選択" return-object item-text="name" item-value="id" append-outer-icon="add" @click:append-outer="appendPriorityCircle" :disabled="editPriorities.length >= 5")

        v-card-actions
          v-spacer
          v-btn(flat @click.native="priorityDialog = false" :disabled="sending") キャンセル
          v-btn(flat color="primary" :disabled="sending" :loading="sending" @click.native="updatePriority") OK

</template>

<script>
import api from '../api'
import RequestCircleItemList from '../components/RequestCircleItemList'

export default {
  name: 'MyRequests',
  components: {
    RequestCircleItemList
  },
  data: function () {
    return {
      selectedDay: '１日目',
      daySelectItems: ['企業', '１日目', '２日目', '３日目', '全日'],
      requests: [],
      requestMap: new Map(),
      requestCircleMap: new Map(),
      priorityData: {
        enterprise: [],
        day1: [],
        day2: [],
        day3: []
      },
      dialog: false,
      priorityDialog: false,
      sending: false,
      deleting: false,
      editing: false,
      editTarget: null,
      editTargetCircle: null,
      editNum: 1,
      editPrice: -1,
      editPriorities: [],
      editPrioritySelectCircle: null,
      rules: {
        required: value => !!value || '必須項目です'
      },
      valid: false
    }
  },
  computed: {
    numEditable: function () {
      if (this.editTarget == null) {
        return false
      }
      switch (this.requestCircleMap.get(this.editTarget.item.circle_id).circle.day) {
        case 0:
          return !this.$store.getters.isEnterpriseDeadlineOver
        case 1:
          return !this.$store.getters.isDay1DeadlineOver
        case 2:
          return !this.$store.getters.isDay2DeadlineOver
        case 3:
          return !this.$store.getters.isDay3DeadlineOver
        default:
          return false
      }
    },
    filteredRequests: function () {
      const day = this.selectedDayNum
      if (day == null) return this.requests
      const c = this.requests.filter((v, i, a) => v.circle.day === day)
      let d
      switch (day) {
        case 0:
          if (this.priorityData.enterprise == null) return c
          d = this.priorityData.enterprise
          break
        case 1:
          if (this.priorityData.day1 == null) return c
          d = this.priorityData.day1
          break
        case 2:
          if (this.priorityData.day2 == null) return c
          d = this.priorityData.day2
          break
        case 3:
          if (this.priorityData.day3 == null) return c
          d = this.priorityData.day3
          break
      }

      c.sort((a, b) => {
        let ap = 5 - d.indexOf(a.circle_id)
        if (ap > 5) {
          ap = -1
        }
        let bp = 5 - d.indexOf(b.circle_id)
        if (bp > 5) {
          bp = -1
        }

        return bp - ap
      })
      return c
    },
    totalCosts: function () {
      let unknown = false
      let costs = 0
      for (const circle of this.filteredRequests) {
        for (const req of circle.items) {
          if (req.item.price === -1) {
            unknown = true
          } else {
            costs += req.item.price * req.num
          }
        }
      }

      if (unknown) {
        return `${costs}円 + α`
      } else {
        return `${costs}円`
      }
    },
    circleCount: function () {
      return this.filteredRequests.length
    },
    prioritySelectItems: function () {
      return this.filteredRequests.map(v => v.circle).filter(v => this.editPriorities.every(w => v.id !== w.id))
    },
    selectedDayNum: function () {
      switch (this.selectedDay) {
        case '企業':
          return 0
        case '１日目':
          return 1
        case '２日目':
          return 2
        case '３日目':
          return 3
        default:
          return null
      }
    },
    isSelectedDayDeadlineOver: function () {
      switch (this.selectedDayNum) {
        case 0:
          return this.$store.getters.isEnterpriseDeadlineOver
        case 1:
          return this.$store.getters.isDay1DeadlineOver
        case 2:
          return this.$store.getters.isDay2DeadlineOver
        case 3:
          return this.$store.getters.isDay3DeadlineOver
        default:
          return true
      }
    }
  },
  created: async function () {
    await this.reloadMyRequests()
    if (this.$route.query.day) {
      switch (this.$route.query.day) {
        case '0':
          this.selectedDay = '企業'
          break
        case '1':
          this.selectedDay = '１日目'
          break
        case '2':
          this.selectedDay = '２日目'
          break
        case '3':
          this.selectedDay = '３日目'
          break
      }
    }
  },
  methods: {
    reloadMyRequests: async function () {
      try {
        const priorities = await api.getMyPriority()
        this.priorityData.enterprise = priorities.enterprise
        this.priorityData.day1 = priorities.day1
        this.priorityData.day2 = priorities.day2
        this.priorityData.day3 = priorities.day3

        const data = await api.getMyRequests()
        this.requestCircleMap.clear()
        this.requestMap.clear()
        for (const circle of data) {
          circle.circle = await api.getCircle(circle.circle_id)
          this.requestCircleMap.set(circle.circle_id, circle)
          for (const request of circle.items) {
            this.requestMap.set(request.id, request)
          }
        }
        this.requests = data
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    openEditDialog: function (reqId) {
      const req = this.requestMap.get(reqId)
      if (typeof req === 'undefined') return
      this.editTarget = req
      this.editNum = req.num
      this.editPrice = req.item.price === -1 ? '' : req.item.price
      this.dialog = true
    },
    editRequest: async function () {
      this.sending = true
      this.editing = true
      try {
        const p = this.editPrice === '' ? -1 : this.editPrice
        if (this.editTarget.num !== this.editNum) {
          await api.editRequest(this.editTarget.id, this.editNum)
        }
        if (this.editTarget.item.price !== p) {
          await api.patchCircleItemPrice(this.editTarget.item_id, p)
        }

        this.editTarget.num = this.editNum
        this.editTarget.item.price = p
        this.dialog = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.editing = false
      this.sending = false
    },
    deleteRequest: async function () {
      this.sending = true
      this.deleting = true
      try {
        await api.deleteRequest(this.editTarget.id)

        this.requestMap.delete(this.editTarget.id)
        for (const circle of this.requests) {
          if (circle.circle_id === this.editTarget.item.circle_id) {
            let i = 0
            while (circle.items[i].id !== this.editTarget.id) i++
            circle.items.splice(i, 1)
          }
        }

        this.dialog = false
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.deleting = false
      this.sending = false
    },
    openEditPriorityDialog: async function () {
      try {
        let ids
        switch (this.selectedDayNum) {
          case 0:
            ids = this.priorityData.enterprise
            break
          case 1:
            ids = this.priorityData.day1
            break
          case 2:
            ids = this.priorityData.day2
            break
          case 3:
            ids = this.priorityData.day3
            break
          default:
            return
        }

        this.editPriorities.splice(0)
        if (ids != null) {
          for (const id of ids) {
            this.editPriorities.push(await api.getCircle(id))
          }
        }
        this.priorityDialog = true
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    appendPriorityCircle: function () {
      if (this.editPrioritySelectCircle == null) return
      this.editPriorities.push(this.editPrioritySelectCircle)
      this.editPrioritySelectCircle = null
    },
    removePriorityCircle: function (idx) {
      this.editPriorities.splice(idx, 1)
    },
    updatePriority: async function () {
      this.sending = true
      try {
        await api.setMyPriority(this.selectedDayNum, this.editPriorities.map(v => v.id))
        switch (this.selectedDayNum) {
          case 0:
            this.priorityData.enterprise = this.editPriorities.map(v => v.id)
            break
          case 1:
            this.priorityData.day1 = this.editPriorities.map(v => v.id)
            break
          case 2:
            this.priorityData.day2 = this.editPriorities.map(v => v.id)
            break
          case 3:
            this.priorityData.day3 = this.editPriorities.map(v => v.id)
            break
        }
        this.priorityDialog = false
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
}
</script>
