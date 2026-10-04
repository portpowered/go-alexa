GO ?= go
GOLANGCI_LINT_VERSION := v2.3.0
GOLANGCI_LINT ?= $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
CLI_DIR := cmd/go-alexa

export GOWORK := off

.DEFAULT_GOAL := check
.PHONY: check format-check vet build test lint fmt cli-check cli-vet cli-build cli-test cli-lint cli-fmt module cli-module generate generate-graphql generate-api model-inventory openapi-bundle openapi-bundle-check

check: format-check lint vet build test model-inventory openapi-bundle-check module cli-module

ifeq ($(OS),Windows_NT)
format-check:
	@powershell -NoProfile -Command "$$files = git ls-files -- '*.go'; if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; $$unformatted = gofmt -l $$files; if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; if ($$unformatted) { Write-Output 'Unformatted Go files:'; $$unformatted; exit 1 }"
else
format-check:
	@unformatted="$$(gofmt -l $$(git ls-files -- '*.go'))" || exit $$?; \
	if [ -n "$$unformatted" ]; then echo "Unformatted Go files: $$unformatted"; exit 1; fi
endif
vet:
	$(GO) vet ./...
	cd $(CLI_DIR) && $(GO) vet ./...

build:
	$(GO) build ./...
	cd $(CLI_DIR) && $(GO) build ./...

test:
	$(GO) test -race ./...
	cd $(CLI_DIR) && $(GO) test -race ./...

lint:
	$(GOLANGCI_LINT) run --config .golangci.yml ./...
	cd $(CLI_DIR) && $(GOLANGCI_LINT) run --config .golangci.yml ./...

fmt:
	$(GO) fmt ./...
	cd $(CLI_DIR) && $(GO) fmt ./...

cli-vet:
	$(GO) -C $(CLI_DIR) vet ./...

cli-build:
	cd $(CLI_DIR) && $(GO) build ./...

cli-test:
	cd $(CLI_DIR) && $(GO) test -race ./...

cli-lint:
	cd $(CLI_DIR) && $(GOLANGCI_LINT) run --config .golangci.yml ./...

cli-fmt:
	cd $(CLI_DIR) && $(GO) fmt ./...

module:
	$(GO) mod tidy -diff
	$(GO) mod verify

cli-module:
	cd $(CLI_DIR) && $(GO) mod tidy -diff && $(GO) mod verify

cli-check: cli-lint cli-vet cli-build cli-test cli-module

generate: generate-graphql generate-api
	$(GO) run ./tools/modelinventory -write

model-inventory:
	$(GO) run ./tools/modelinventory -check

openapi-bundle:
	$(GO) run ./tools/openapibundle -write

openapi-bundle-check:
	$(GO) run ./tools/openapibundle -check

generate-graphql:
	$(GO) tool genqlient
	$(GO) run ./tools/graphqlops
	$(GO) run ./tools/graphqlmodels

generate-api: openapi-bundle
	$(GO) run ./tools/wireconstants
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/account-linking.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/auth-compat.config.yaml api/compat/auth.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/endpoints.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/endpoints-compat.config.yaml api/compat/endpoints.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/devices-v2.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/devices-compat.config.yaml api/compat/devices.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/user-info.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/user-info-compat.config.yaml api/compat/user-info.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/behaviors.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/behaviors-compat.config.yaml api/compat/behaviors.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/controls-compat.config.yaml api/compat/controls.yaml
	$(GO) run ./tools/oapiallOfix
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/graphql-transport.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/media.config.yaml api/openapi.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/media-compat.config.yaml api/compat/media.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/events-compat.config.yaml api/compat/events.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/behavior-sequence.config.yaml api/behaviors.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/feature-events.config.yaml api/feature-events.yaml
	$(GO) run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config pkg/dependencymodels/feature-controls.config.yaml api/feature-controls.yaml
	cd tools/protocols && npm ci --ignore-scripts && node prepare-generated.mjs && node generate-directives.mjs && node stamp-generated.mjs
	$(GO) run ./tools/apiroutes
	$(GO) fmt ./pkg/dependencymodels ./pkg/internal/apiroutes
