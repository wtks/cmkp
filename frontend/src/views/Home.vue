<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(column)
      v-flex
        v-card
          v-card-text
            span メニューの「マイリクエスト」または「サークル検索」から、買ってきて欲しい物のリクエスト等を行ってください。
            br
            span リクエストをし終わったら、必ず「マイリクエスト」の「希望優先順位を設定する」から特に優先して買ってきて欲しいサークルを、日程毎に設定してください。
            br
            span リクエストに関して管理人に伝えたいことがある場合は、メニューの「リクエスト備考」からリクエスト備考を書くことができます。
      v-flex
        v-card
          v-card-title リクエスト受付締め切り日時
          v-card-text
            p 企業: {{ enterpriseDeadline.format('MM/DD HH:mm') }}
            p 1日目: {{ day1Deadline.format('MM/DD HH:mm') }}
            p 2日目: {{ day2Deadline.format('MM/DD HH:mm') }}
            p 3日目: {{ day3Deadline.format('MM/DD HH:mm') }}

</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'home',
  computed: {
    ...mapGetters([
      'enterpriseDeadline',
      'day1Deadline',
      'day2Deadline',
      'day3Deadline'
    ])
  }
}
</script>
