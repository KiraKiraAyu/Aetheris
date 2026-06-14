.PHONY: test tidy run-api run-worker web-dev web-build

test:
	cd core && go test ./...

tidy:
	cd core && go mod tidy

run-api:
	cd core && go run ./cmd/api

run-worker:
	cd core && go run ./cmd/worker

web-dev:
	cd web && pnpm run dev --host 127.0.0.1 --port 5178 --strictPort

web-build:
	cd web && pnpm run build
