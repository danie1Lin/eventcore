.PHONY: help
help:## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: example
example:## Run example
	go run example/main.go

.PHONY: cleanup
cleanup:## Clean up rabbitmq's queue and exchange after run example
	if ! [ -d tmp ];then mkdir tmp; fi;
	- rabbitmqctl list_queues > tmp/del_q
	- cat tmp/del_q | awk '/main.EventTest.*/ {system("rabbitmqctl delete_queue " $$1)}'
	rm tmp/del_q
	rabbitmqadmin delete exchange name='main.EventTest'