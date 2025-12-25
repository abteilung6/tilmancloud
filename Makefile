.PHONY: generate-go-models generate-ts-client validate-spec frontend-format frontend-lint frontend-lint-fix frontend-type-check frontend-check run-admin-api run-frontend

OAPI_CODEGEN := $(shell which oapi-codegen || echo $(HOME)/go/bin/oapi-codegen)

validate-spec:
	@echo "Validating OpenAPI spec..."
	@if [ ! -f "$(OAPI_CODEGEN)" ]; then \
		echo "oapi-codegen not found. Install with: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest"; \
		exit 1; \
	fi

generate-go-models: validate-spec
	@echo "Generating Go models from OpenAPI spec..."
	@mkdir -p pkg/api/generated
	@$(OAPI_CODEGEN) -generate types \
		-package generated \
		-o pkg/api/generated/models.go \
		api/openapi.yaml
	@echo "✓ Go models generated successfully"

# Frontend formatting and linting
frontend-format:
	@echo "Formatting frontend code..."
	@cd console && npm run format
	@echo "✓ Formatting complete"

frontend-lint:
	@echo "Linting frontend code..."
	@cd console && npm run lint
	@echo "✓ Linting complete"

frontend-lint-fix:
	@echo "Fixing linting issues..."
	@cd console && npm run lint:fix
	@echo "✓ Lint fixes applied"

frontend-type-check:
	@echo "Type checking frontend code..."
	@cd console && npm run type-check
	@echo "✓ Type check complete"

frontend-check: frontend-type-check frontend-lint
	@echo "Running format check..."
	@cd console && npm run format:check
	@echo "✓ All frontend checks passed"

# TypeScript API Client Generation
generate-ts-client:
	@echo "Generating TypeScript API client from OpenAPI spec..."
	@cd console && npm run generate-api
	@echo "✓ TypeScript API client generated successfully"

# Server commands
run-admin-api:
	@echo "Starting API server on http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@go run cmd/admin-api/main.go

run-frontend:
	@echo "Starting frontend development server..."
	@echo "Press Ctrl+C to stop"
	@cd console && npm run dev


