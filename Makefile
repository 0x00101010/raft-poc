.PHONY: build
build:
	go build -o bin/leader_elector ./cmd/main.go

.PHONY: bootstrap
bootstrap:
	scripts/bootstrap.sh

.PHONY: run
run:
	docker-compose up