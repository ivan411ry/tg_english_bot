
run:
	docker compose up --build english_bot

up:
	docker compose up -d postgres_english_bot

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres_english_bot:5432/english_bot?sslmode=disable" down 1

migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=create_users"; exit 1; fi
	docker compose run --rm migrate create -ext sql -dir /migrations -seq $(name)

down:
	docker compose down

env:
	cp .env.example .env