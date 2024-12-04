#postgres:
Run_postgres:
	docker run --name  postgres_urls  -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine


Run_redis:
	docker run --name redis_urls -p 6379:6379 -d redis



migrateup:
	migrate -path ./database/migrate -database "postgresql://root:secret@localhost:5432/urldb?sslmode=disable" -verbose up


migratedown:
	migrate -path ./database/migrate -database "postgresql://root:secret@localhost:5432/urldb?sslmode=disable" -verbose down
