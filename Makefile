help:## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

example:## run example
	go run example/main.go

cleanup:## clean up queue and exchange after run example
	rabbitmqctl list_queues > del_q
	cat del_q | awk '/main.EventTest.*/ {system("rabbitmqctl delete_queue " $$1)}'
	rabbitmqadmin delete exchange name='main.EventTest'