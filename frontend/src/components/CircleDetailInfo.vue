<template lang="pug">
  v-card
    v-card-title.headline サークル情報
    v-card-text
      template(v-if="loading") 読み込み中
      dl(v-else)
        dt サークル名
        dd {{ circle.name }}
        dt 作家名
        dd {{ circle.author }}
        dt 場所
        dd
          | {{ locationString }}
          span.orange--text(v-if="circle.location_type === 1") {{ locationTypeString }}
          span.red--text(v-else-if="circle.location_type === 2") {{ locationTypeString }}
        dt ジャンル
        dd {{ circle.genre }}
        template()
          dt 外部リンク
          dd
            v-btn(v-if="circle.website != null" icon :href="circle.website" target="_blank")
              v-icon home
            v-btn(v-if="circle.twitter_id != null" icon :href="'https://twitter.com/'+circle.twitter_id" target="_blank")
              v-icon(color="blue") fab fa-twitter
            v-btn(v-if="circle.pixiv_id != null" icon :href="'https://www.pixiv.net/member.php?id='+circle.pixiv_id" target="_blank")
              img(src="../assets/pixiv_icon.jpg" height="24px" width="24px")
            v-btn(v-if="circle.niconico_id != null" small flat :href="'http://www.nicovideo.jp/user/'+circle.niconico_id" target="_blank") Niconico

</template>

<script>
import api from '../api'

export default {
  name: 'CircleDetailInfo',
  data: function () {
    return {
      loading: true,
      circle: null
    }
  },
  props: {
    id: {
      type: Number,
      required: true
    }
  },
  created: async function () {
    await this.updateCircle()
  },
  watch: {
    id: async function () {
      await this.updateCircle()
    }
  },
  computed: {
    locationString: function () {
      return this.circle.day !== 0 ? this.circle.hall + this.circle.block + this.circle.space : this.circle.hall + this.circle.space
    },
    locationTypeString: function () {
      switch (this.circle.location_type) {
        case 0:
          return ''
        case 1:
          return '壁'
        case 2:
          return 'シャッター'
        default:
          return ''
      }
    }
  },
  methods: {
    updateCircle: async function () {
      this.loading = true
      if (this.id == null) {
        return
      }
      try {
        this.circle = await api.getCircle(this.id)
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
      this.loading = false
    }
  }
}
</script>
