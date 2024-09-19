tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
	go install github.com/volatiletech/sqlboiler/v4@v4.13.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.13.0

migrate-up:
	migrate -database 'postgres://ants_celestia:password@localhost:5432/ants_celestia?sslmode=disable' -path db/migrations up

migrate-down:
	migrate -database 'postgres://ants_celestia:password@localhost:5432/ants_celestia?sslmode=disable' -path db/migrations down

models:
	sqlboiler --no-tests psql