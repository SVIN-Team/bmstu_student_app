
GO_BUILD := go build
GO_TEST := go test
SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6
EXE_PATH := build/app

$(EXE_PATH): cmd/main.go
	mkdir -p $(dir $@)
	$(GO_BUILD) -o $@ $<

.PHONY: build
build: $(EXE_PATH)

.PHONY: run
run: $(EXE_PATH)
	ACCESS_SECRET_KEY=dev REFRESH_SECRET_KEY=dev $< -c ../deploy/config/config.yaml

.PHONY: clean
clean:
	rm -rf build

.PHONY: unit_test
unit_test:
	$(GO_TEST) -tags=unit ./...

.PHONY: swagger
swagger:
	$(SWAG) init -g cmd/main.go -o docs --parseInternal
