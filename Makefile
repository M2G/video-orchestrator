ifeq ($(CI_COMMIT_REF_NAME),)
    branch = $(shell git rev-parse --abbrev-ref HEAD)
else
    branch = tag:$(CI_COMMIT_REF_NAME)
endif

commit = $(shell git log --pretty=format:'%H' -n 1)
now = $(shell date "+%Y-%m-%d %T UTC%z")
compiler = $(shell go version)


IMAGE_NAME :=  registry.github.com/video-orchestrator

all: test build image
.PHONY: all test clean

test:
	@echo "Running tests"
	@docker-compose -f docker-compose.test.yml up	\
	--build											\
	--abort-on-container-exit						\
	--force-recreate								\
	--quiet-pull									\
	--no-color										\
	--remove-orphans								\
	--timeout 100
	@docker-compose -f docker-compose.test.yml rm -f

build:
	@echo "Compiling the binaries"
	CGO_ENABLED=0									\
	GOBIN=$(PWD)/bin								\
	go install  -v									\
	-ldflags										\
	"-X 'main.branch=$(branch)'						\
	-X 'main.sha=$(commit)'							\
	-X 'main.compiledAt=$(now)'						\
	-X 'main.compiler=$(compiler)'					\
	-s -w"											\
	-a -installsuffix cgo ./...

image:
	@echo "Building Docker Image"
	@(docker build -t $(IMAGE_NAME) .)


clean:
	@rm -rf $(PWD)/bin/*

re: clean build

dev:
	@echo "Running dev"
	@docker-compose -f docker-compose.dev.yml up	\
		--force-recreate							\
		--quiet-pull								\
		--no-color									\
		--remove-orphans
	@docker-compose rm -f
