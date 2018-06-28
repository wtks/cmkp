/* eslint-disable no-template-curly-in-string */
module.exports = {
  'presets': [
    [
      '@vue/app',
      {
        'useBuiltIns': 'entry'
      }
    ]
  ],
  'plugins': [
    [
      'transform-imports',
      {
        'vuetify': {
          'transform': 'vuetify/es5/components/${member}',
          'preventFullImport': true
        }
      }
    ]
  ]
}
