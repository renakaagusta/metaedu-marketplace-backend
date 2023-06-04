# alias sqlc='$(go env GOPATH)/bin/sqlc'
# alias air='$(go env GOPATH)/bin/air'

dev:
	docker-compose up -d

dev-down:
	docker-compose down

psql:
	psql --host=127.0.0.1 --port=6500 --username=admin --dbname=metaedu -W

truncate:
	TRUNCATE TABLE tokens; TRUNCATE TABLE rentals; TRUNCATE TABLE ownerships;  TRUNCATE TABLE transactions; TRUNCATE TABLE fractions; 
	
go:
	air

migrate:
	migrate create -ext sql -dir db/migrations -seq init_schema

migrate-up:
	migrate -path db/migrations -database "postgresql://admin:password123@localhost:6500/metaedu?sslmode=disable" -verbose up

migrate-down:
	migrate -path db/migrations -database "postgresql://admin:password123@localhost:6500/metaedu?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: sqlc