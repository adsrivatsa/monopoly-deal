postgresup:
	docker run --name postgres -p 5432:5432 -e PGUSER=postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=wealth-management -d postgres

postgresdown:
	@docker stop postgres || true
	@docker rm postgres || true

createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres monopoly-deal

dropdb:
	docker exec -it postgres dropdb monopoly-deal

migrateup:
	migrate -path internal/migration -database "postgresql://postgres:password@localhost:5432/monopoly-deal?sslmode=disable" -verbose up

migratedown: postgresdown postgresup
	# migrate -path internal/migration -database "postgresql://postgres:password@localhost:5432/monopoly-deal?sslmode=disable" -verbose down