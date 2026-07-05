# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./cmd/main.go
BINARY_NAME := go-webservice

# git tag or commit hash
SERVICE_VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# ==================================================================================== #
# Dependencies
# ==================================================================================== #

## otel/start: start the otel trace collector stack
.PHONY: otel/start
otel/start:
	@podman compose up -d

## otel/stop: stop the otel trace collector stack
.PHONY: otel/stop
otel/stop:
	@podman compose down

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=./build/coverage.out ./...
	go tool cover -html=./build/coverage.out

## bdd: run bdd tests
.PHONY: bdd
bdd:
	go test -v -race -buildvcs ./test/... -tags integration

## build: build the application
.PHONY: build
build:
	go build -ldflags="-X main.serviceVersion=${SERVICE_VERSION} -X main.gitCommit=${GIT_COMMIT}" -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## run: run the  application
.PHONY: run
run: build
	/tmp/bin/${BINARY_NAME} --port 9090

## run/local: run the application and any local dependencies as test containers
.PHONY: run/local
run/local:
	go run -tags local ./cmd

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
        --build.cmd "make build" --build.bin "/tmp/bin/${BINARY_NAME}" --build.delay "100" \
        --build.exclude_dir "test" \
        --build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
        --misc.clean_on_exit "true"


# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## docker/build: build the docker image
.PHONY: docker/build
docker/build:
	docker build -t ${BINARY_NAME} .

## docker/run: run the application via docker
.PHONY: docker/run
docker/run: docker/build
	docker run -it --rm -p 9090:9090 ${BINARY_NAME}

## push: push changes to the remote Git repository
.PHONY: push
push: tidy audit no-dirty
	git push

## production/deploy: deploy the application to production
.PHONY: production/deploy
production/deploy: confirm tidy audit no-dirty
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -X main.serviceVersion=${SERVICE_VERSION} -X main.gitCommit=${GIT_COMMIT}' -o=/tmp/bin/linux_amd64/${BINARY_NAME} ${MAIN_PACKAGE_PATH}
	upx -5 /tmp/bin/linux_amd64/${BINARY_NAME}
    # Include additional deployment steps here...