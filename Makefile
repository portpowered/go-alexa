GO ?= go

.DEFAULT_GOAL := check
.PHONY: check build test lint fmt generate generate-graphql generate-api

check: lint build test

build:
	$(GO) build ./...

test:
	$(GO) test -race ./...

lint:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

generate: generate-graphql generate-api

generate-graphql:
	$(GO) tool genqlient
	$(GO) run ./tools/graphqlops

generate-api:
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencies/internal/wire/config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/embedded-wire.config.yaml api/embedded-wire.yaml
	cd tools/protocols && npm ci --ignore-scripts && node -e "const fs=require('fs'); const dir='../../pkg/alexa/internal/wire'; for (const file of fs.readdirSync(dir)) if (file.endsWith('.go')) fs.unlinkSync(dir+'/'+file)" && npx --no-install modelina generate golang ../../api/asyncapi.yaml --packageName wire --goIncludeTags -o ../../pkg/alexa/internal/wire
	$(GO) run ./tools/apiroutes
	$(GO) fmt ./pkg/dependencies/internal/wire ./pkg/dependencymodels ./pkg/alexa/internal/wire ./pkg/internal/apiroutes
