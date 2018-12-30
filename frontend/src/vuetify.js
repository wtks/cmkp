import Vue from 'vue'
import {
  transitions,
  VAlert,
  VApp,
  VAutocomplete,
  VBtn,
  VCard,
  VCheckbox,
  VChip,
  VDataIterator,
  VDatePicker,
  VDialog,
  VDivider,
  VFooter,
  VForm,
  VGrid,
  VIcon,
  VList,
  VNavigationDrawer,
  VProgressLinear,
  VRadioGroup,
  VSelect,
  VSpeedDial,
  VSubheader,
  VTextarea,
  VTextField,
  VTimePicker,
  VToolbar,
  Vuetify
} from 'vuetify'
import 'vuetify/src/stylus/app.styl'
import { Scroll } from 'vuetify/es5/directives'

Vue.use(Vuetify, {
  components: {
    VApp,
    VNavigationDrawer,
    VFooter,
    VList,
    VBtn,
    VIcon,
    VGrid,
    VToolbar,
    VTextField,
    VCard,
    VForm,
    VDialog,
    VDivider,
    VDataIterator,
    VCheckbox,
    VTextarea,
    VChip,
    VSelect,
    VRadioGroup,
    VTimePicker,
    VDatePicker,
    VSubheader,
    VAutocomplete,
    VSpeedDial,
    VProgressLinear,
    VAlert,
    transitions
  },
  directives: {
    Scroll
  }
})
