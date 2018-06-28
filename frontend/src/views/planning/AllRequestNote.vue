<template lang="pug">
  v-container(fluid grid-list-md)
    v-layout(column)
      v-flex(xs12 v-for="noteId in notes" :key="noteId")
        v-card
          v-card-text
            request-note(:id="noteId")

</template>

<script>
import api from '../../api'
import RequestNote from '../../components/RequestNote'

export default {
  name: 'AllRequestNote',
  components: {
    RequestNote
  },
  data: function () {
    return {
      notes: []
    }
  },
  mounted: async function () {
    await this.reload()
  },
  methods: {
    reload: async function () {
      try {
        this.notes = await api.getAllRequestNotes()
      } catch (e) {
        console.error(e)
        if (e.response) {
          this.$bus.$emit('error', e.response.data.message)
        } else {
          this.$bus.$emit('error')
        }
      }
    }
  }
}
</script>
