include ./.env
DBURL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATION_PATH=db/migrations
SEED_PATH=db/seeds

migrate-create:
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq create_$(NAME)_table

migrate-up:
	migrate -database $(DBURL) -path $(MIGRATION_PATH) up

insert-seed:
	for file in $$(ls $(SEED_PATH)/*.sql | sort); do \
		echo "Seeding $$file..."; \
		psql "$(DBURL)" -f $$file; \
	done

migrate-down:
	migrate -database $(DBURL) -path $(MIGRATION_PATH) down
