.PHONY: all build test lint migrate install run stop

all: build

install:
	./install.sh

run:
	./run_all.sh

stop:
	./stop_all.sh

build:
	cd backend && make build

test:
	cd backend && make test
	cd web && pnpm test
	cd mobile && melos run test

lint:
	cd backend && make lint
	cd web && pnpm lint
	cd mobile && melos run analyze

migrate:
	cd backend && make migrate
