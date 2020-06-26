import Vuex from 'vuex'
import eventCenter from './event_center/client'

const debug = process.env.NODE_ENV !== 'production'

export default new Vuex.Store({
	modules: {
		eventCenter,
	},
	strict: debug,
	//plugins: debug ? [createLogger()] : [],
})
