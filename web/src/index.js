import EventCenter from './event_center_client'

var client = new EventCenter('localhost', '8080')
client
	.subscript(
		{ type: 'main.EventTest' },
		(e) => {
			console.log('f1', e)
		},
		(e) => {
			console.log('f2', e, e)
		}
	)
	.then(() => {
		client.eventTunnel()

		console.log('')
	})
