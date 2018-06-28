<template lang="pug">
  v-container(fluid grid-list-md)
    span {{ filteredRequestedCircleCount }}サークル (壁:{{filteredRequestedWallCircleCount}}, シャッター:{{filteredRequestedShutterCircleCount}})
    v-layout(row wrap)
      v-flex
        v-radio-group(v-model="filter.day" row)
          v-radio(label="企業" :value="0")
          v-radio(label="1日目" :value="1")
          v-radio(label="2日目" :value="2")
          v-radio(label="3日目" :value="3")
      v-flex
        v-checkbox(label="通常" v-model="filter.normal")
      v-flex
        v-checkbox(label="壁" v-model="filter.wall")
      v-flex
        v-checkbox(label="シャッター" v-model="filter.shutter")

    v-layout(row wrap)
      v-flex(xs12 sm12 md6 lg4 v-for="circle in filteredRequests" :key="circle.circle_id")
        planner-request-circle-item-list(v-bind="circle" @itemClick="onItemClicked")
    v-dialog(v-model="itemDetail.dialog" width=500)
      planner-request-circle-item-detail(v-if="itemDetailValid" v-bind="itemDetail")

</template>

<script>
import api from '../../api'
import PlannerRequestCircleItemList from '../../components/PlannerRequestCircleItemList'
import PlannerRequestCircleItemDetail from '../../components/PlannerRequestCircleItemDetail'

export default {
  name: 'AllRequestList',
  components: {
    PlannerRequestCircleItemList,
    PlannerRequestCircleItemDetail
  },
  data: function () {
    return {
      requests: [],
      circleMap: new Map(),
      itemMap: new Map(),
      requestMap: new Map(),
      itemDetail: {
        dialog: false,
        circleId: null,
        circle: null,
        itemId: null,
        item: null
      },
      filter: {
        day: 1,
        normal: true,
        wall: true,
        shutter: true
      }
    }
  },
  computed: {
    filteredRequests: function () {
      return this.requests.filter(v => {
        let ok = true
        if (this.filter.day != null) {
          ok = v.circle.day === this.filter.day
        }
        if (!ok) return false
        switch (v.circle.location_type) {
          case 0:
            ok = this.filter.normal
            break
          case 1:
            ok = this.filter.wall
            break
          case 2:
            ok = this.filter.shutter
            break
        }
        return ok
      })
    },
    filteredRequestedCircleCount: function () {
      return this.filteredRequests.length
    },
    filteredRequestedWallCircleCount: function () {
      return this.filteredRequests.reduce((x, y) => x + (y.circle.location_type === 1 ? 1 : 0), 0)
    },
    filteredRequestedShutterCircleCount: function () {
      return this.filteredRequests.reduce((x, y) => x + (y.circle.location_type === 2 ? 1 : 0), 0)
    },
    itemDetailValid: function () {
      return !!this.itemDetail.circle && !!this.itemDetail.item
    }
  },
  watch: {
    'itemDetail.circleId': function () {
      this.itemDetail.circle = this.circleMap.get(this.itemDetail.circleId)
    },
    'itemDetail.itemId': function () {
      this.itemDetail.item = this.itemMap.get(this.itemDetail.itemId)
    }
  },
  mounted: async function () {
    await this.reloadRequests()
  },
  methods: {
    reloadRequests: async function () {
      try {
        const data = await api.getAllRequests()
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
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    },
    onItemClicked: function ({circleId, itemId}) {
      this.itemDetail.circleId = circleId
      this.itemDetail.itemId = itemId
      this.itemDetail.dialog = true
    }
  }
}
</script>
