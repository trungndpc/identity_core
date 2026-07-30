.PHONY: run deps setup-db migrate

deps:
	go mod tidy

run:
	go run ./cmd/server

# Tạo database local (chạy 1 lần, cần PostgreSQL đã cài sẵn trên máy)
setup-db:
	@psql -U postgres -f scripts/setup_db.sql 2>/dev/null \
		|| psql -f scripts/setup_db.sql

migrate:
	@bash scripts/run_migrations.sh
