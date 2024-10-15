tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
	go install github.com/volatiletech/sqlboiler/v4@v4.13.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.13.0

models:
	sqlboiler --no-tests psql

migrate-up:
	migrate -database 'postgres://ants_watch:password@localhost:5432/ants_watch?sslmode=disable' -path db/migrations up

migrate-down:
	migrate -database 'postgres://ants_watch:password@localhost:5432/ants_watch?sslmode=disable' -path db/migrations down
