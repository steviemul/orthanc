.DEFAULT_GOAL := build-observer

OBSERVER_DIR := client/observer
DOCKER_DIR := docker

.PHONY: build-observer docker-up docker-down clean help

all: help

build-observer:
	$(MAKE) -C $(OBSERVER_DIR)

docker-up:
	docker compose -f $(DOCKER_DIR)/docker-compose.yml up --build

docker-down:
	docker compose -f $(DOCKER_DIR)/docker-compose.yml down

clean:
	$(MAKE) -C $(CLIENT_DIR) clean