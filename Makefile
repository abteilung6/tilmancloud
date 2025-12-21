.PHONY: generate-models validate-spec

OAPI_CODEGEN := $(shell which oapi-codegen || echo $(HOME)/go/bin/oapi-codegen)

validate-spec:
	@echo "Validating OpenAPI spec..."
	@if [ ! -f "$(OAPI_CODEGEN)" ]; then \
		echo "oapi-codegen not found. Install with: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest"; \
		exit 1; \
	fi

generate-openapi-models: validate-spec
	@echo "Generating Go models from OpenAPI spec..."
	@mkdir -p pkg/api/generated
	@$(OAPI_CODEGEN) -generate types \
		-package generated \
		-o pkg/api/generated/models.go \
		api/openapi.yaml
	@echo "✓ Models generated successfully"
