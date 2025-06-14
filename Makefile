.PHONY: mocks
mocks:
	mockgen -source=internal/repository/user_repository.go -destination=internal/repository/mocks/user_repository_gomock.go -package=mocks
	mockgen -source=internal/repository/appointment_repository.go -destination=internal/repository/mocks/appointment_repository_gomock.go -package=mocks

.PHONY: test-unit
test-unit: mocks
	go test ./internal/... -v

.PHONY: test-integration
test-integration:
	RUN_INTEGRATION_TESTS=true go test ./cmd/... -v