.DEFAULT_GOAL := build-observer

OBSERVER_DIR := client/observer
EVENT_COLLECTOR_DIR := server/event-collector
DOCKER_DIR := docker

.PHONY: build-observer build-event-collector docker-up docker-down clean help

all: help

build-observer:
	$(MAKE) -C $(OBSERVER_DIR)

build-event-collector:
	bash -lc 'source "$$HOME/.sdkman/bin/sdkman-init.sh" && cd $(EVENT_COLLECTOR_DIR) && sdk env && ./mvnw -B package'

docker-up:
	docker compose -f $(DOCKER_DIR)/docker-compose.yml up --build

docker-down:
	docker compose -f $(DOCKER_DIR)/docker-compose.yml down

clean:
	$(MAKE) -C $(CLIENT_DIR) clean