<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(row wrap v-if="cid != null")
      v-flex
        circle-detail-info(:id="cid")
      v-flex
        circle-memo-list(:id="cid")
    v-btn(block color="success" :disabled="isDeadlineOver" :to="'/my-requests/create/'+cid") リクエストを作成 {{ isDeadlineOver ? '(締め切りました)' : '' }}
</template>

<script>
import CircleDetailInfo from '../components/CircleDetailInfo'
import CircleMemoList from '../components/CircleMemoList'
import api from '../api'

export default {
  name: 'CircleInfo',
  data: function () {
    return {
      cid: null
    }
  },
  components: {
    CircleDetailInfo,
    CircleMemoList
  },
  asyncComputed: {
    isDeadlineOver: async function () {
      if (this.cid == null) {
        return false
      }
      const circle = await api.getCircle(this.cid)
      switch (circle.day) {
        case 0:
          return this.$store.getters.isEnterpriseDeadlineOver
        case 1:
          return this.$store.getters.isDay1DeadlineOver
        case 2:
          return this.$store.getters.isDay2DeadlineOver
        case 3:
          return this.$store.getters.isDay3DeadlineOver
        default:
          return false
      }
    }
  },
  beforeRouteEnter: function (to, from, next) {
    next(vm => {
      vm.cid = parseInt(to.params.cid, 10)
    })
  },
  beforeRouteUpdate: function (to, from, next) {
    this.cid = parseInt(to.params.cid, 10)
    next()
  }
}
</script>
