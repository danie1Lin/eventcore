import Vue from 'vue'
import './plugins/vuex'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import lodash from './plugins/lodash'
import store from './store'
Vue.config.productionTip = false
Vue.use(lodash)

new Vue({
	vuetify,
	store,
	render: (h) => h(App),
}).$mount('#app')
