build:
	@echo "Start go build phase"
	go build -o ./bin/cluster_relay ./main.go
	go build -o ./bin/client_function ./tests/client_function.go