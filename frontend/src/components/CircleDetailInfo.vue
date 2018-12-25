<template lang="pug">
  v-card
    v-card-title.headline サークル情報
    v-card-text
      dl
        dt サークル名
        dd {{ name }}
        dt 作家名
        dd {{ author }}
        dt 場所
        dd
          | {{ locationString }}
          span.orange--text(v-if="locationType === 1") {{ locationTypeString }}
          span.red--text(v-else-if="locationType === 2") {{ locationTypeString }}
        template(v-if="genre != null")
          dt ジャンル
          dd {{ genre }}
        template(v-if="website != null || twitterId != null || pixivId != null || niconicoId != null")
          dt 外部リンク
          dd
            v-btn(v-if="website != null" icon :href="website" target="_blank")
              v-icon home
            v-btn(v-if="twitterId != null" icon :href="`https://twitter.com/${twitterId}`" target="_blank")
              v-icon(color="blue") fab fa-twitter
            v-btn(v-if="pixivId != null" icon :href="`https://www.pixiv.net/member.php?id=${pixivId}`" target="_blank")
              img(src="../assets/pixiv_icon.jpg" height="24px" width="24px")
            v-btn(v-if="niconicoId != null" small flat :href="`http://www.nicovideo.jp/user/${niconicoId}`" target="_blank") Niconico
</template>

<script>
export default {
  name: 'CircleDetailInfo',
  props: {
    id: {
      type: Number,
      required: true
    },
    name: {
      type: String,
      required: true
    },
    author: {
      type: String,
      required: true
    },
    genre: {
      type: String,
      default: null
    },
    day: {
      type: Number,
      required: true
    },
    hall: {
      type: String,
      required: true
    },
    block: {
      type: String,
      required: true
    },
    space: {
      type: String,
      required: true
    },
    locationType: {
      type: Number,
      default: 0
    },
    website: {
      type: String,
      default: null
    },
    twitterId: {
      type: String,
      default: null
    },
    pixivId: {
      type: Number,
      default: null
    },
    niconicoId: {
      type: Number,
      default: null
    }
  },
  computed: {
    locationString: function () {
      return this.day !== 0 ? this.hall + this.block + this.space : this.hall + this.space
    },
    locationTypeString: function () {
      switch (this.locationType) {
        case 0:
          return ''
        case 1:
          return '(壁)'
        case 2:
          return '(シャッター)'
        default:
          return ''
      }
    }
  }
}
</script>
