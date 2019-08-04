module.exports = {
  devServer: {
    proxy: 'http://localhost:5000',
    disableHostCheck: true
  },
  pluginOptions: undefined,
  pwa: {
    workboxOptions: {
      skipWaiting: true,
      clientsClaim: true
    }
  }
}
