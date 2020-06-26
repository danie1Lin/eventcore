import _ from 'lodash'

const VueLodash = {
	install(Vue) {
		Vue.prototype.$_ = _

		Vue.mixin({
			mounted() {
				// Just tell you that it is mounted
				console.log('VueLodash')
			},
		})

		if (typeof window !== 'undefined' && window.Vue) {
			window.Vue.use(VueLodash)
		}
	},
}

export default VueLodash
