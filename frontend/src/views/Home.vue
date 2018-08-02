<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(column)
      v-flex
        v-card
          v-card-text
            span ◎メニューの「マイリクエスト」または「サークル検索」から、買ってきて欲しい物のリクエスト等を行ってください。
            br
            span ◎リクエストをし終わったら、必ず「マイリクエスト」の「希望優先順位を設定する」から特に優先して買ってきて欲しいサークルを、上位５サークルまで日程毎に設定してください。
            br
            span ◎リクエストに関して管理人に伝えたいことがある場合は、メニューの「リクエスト備考」からリクエスト備考を書くことができます。
            br
            span ◎リクエスト受付締め切り後は、リクエストの登録・削除・個数の変更・優先順位設定は出来ません。何かあれば管理人に個別に連絡してください。
            br
            span ◎限数・搬入数やお品書きなどの情報があればそのサークルのサークルメモに書いてもらえると助かります。また、価格未定で登録した商品は価格が決まり次第、リクエストの修正から入力してください。
      v-flex
        v-card
          v-card-title リクエスト受付締め切り日時
          v-card-text
            span 企業: {{ enterpriseDeadline.format('MM/DD HH:mm') }}
            br
            span 1日目: {{ day1Deadline.format('MM/DD HH:mm') }}
            br
            span 2日目: {{ day2Deadline.format('MM/DD HH:mm') }}
            br
            span 3日目: {{ day3Deadline.format('MM/DD HH:mm') }}

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
