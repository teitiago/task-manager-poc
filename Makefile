git-install:  ## install the pre-commit hook
	pre-commit install

coverage-html:  ## build an html coverage report first the tests need to run
	go tool cover -html=coverage.txt

test-all:  ## perform integration and unit tests via docker-compose
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down --volumes --remove-orphans

test-unit:  ## perform unit test validation
	go test ./... --tags=unit

dev-env-up:  ## initialize local dev env
	docker-compose -f docker-compose.yml up --build --abort-on-container-exit

dev-env-down:  ## stop the local env
	docker-compose -f docker-compose.yml down --remove-orphans

dev-env-down-volumes:  ## stop the local env and remove the volumes
	docker-compose -f docker-compose.yml down --volumes --remove-orphans

swagger:  ## generate swagger documentation
	swag init -g internal/server/server.go --output api/docs

lint: golangci gosec

golangci:
	golangci-lint run --allow-parallel-runners --tests=0 ./... 

gosec:
	gosec ./...
