.PHONY: up down del build build-sandboxes logs

up:
	docker compose up -d

down:
	docker compose down

del:
	docker compose down -v

build:
	docker compose build

build-sandboxes:
	docker compose build c-judger cpp-judger python-judger java-judger

logs:
	docker compose logs -f

setup: build build-sandboxes up
	@echo "Judge is up. Sandbox images built."