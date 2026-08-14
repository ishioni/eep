VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

.PHONY: help build build-raw patch-wasm validate-wasm test vet fmt lint clean docker-build docker-run up down logs restart smoke

help:
	@mise tasks

build:
	VERSION=$(VERSION) mise run build

build-raw:
	VERSION=$(VERSION) mise run build-raw

patch-wasm:
	mise run patch-wasm

validate-wasm:
	mise run validate-wasm

test:
	mise run test

vet:
	mise run vet

fmt:
	mise run fmt
	mise run oxfmt

lint:
	mise run lint

clean:
	mise run clean

docker-build:
	VERSION=$(VERSION) mise run docker-build

docker-run:
	VERSION=$(VERSION) mise run docker-run

up: build
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

restart: down up

smoke:
	VERSION=$(VERSION) mise run smoke
