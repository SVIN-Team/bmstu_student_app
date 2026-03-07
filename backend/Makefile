
GO_BUILD := go build
GO_TEST := go test
EXE_PATH := build/app

$(EXE_PATH): cmd/main.go
	$(GO_BUILD) -o $@ $<

.PHONY: build
build: $(EXE_PATH)

.PHONY: run
run: $(EXE_PATH)
	build/main -c ../deploy/config/config.yaml

.PHONY: clean
clean:
	rm -rf build

.PHONY: unit_test
unit_test:
	$(GO_TEST) -tags=unit ./...