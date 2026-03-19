# Build directories
DIST_DIR := dist

# itch.io deployment config (update these)
ITCH_USER := gamekaizen
ITCH_GAME := doomerang

.PHONY: lint run build basic-test \
	build-mac build-mac-intel build-windows build-linux build-web build-all \
	deploy-mac deploy-mac-intel deploy-windows deploy-linux deploy-web deploy-all \
	clean-dist

lint:
	golangci-lint run

vendor:
	go mod vendor

run: vendor
	go run main.go

build:
	go build .

basic-test:
	./scripts/basic-test.sh

# Platform builds
build-mac:
	@mkdir -p $(DIST_DIR)/mac
	CGO_CFLAGS="-w" go build -o $(DIST_DIR)/mac/doomerang .

# build-mac-intel:
# 	@mkdir -p $(DIST_DIR)/mac-intel
# 	CGO_CFLAGS="-w" GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/mac-intel/doomerang .

build-windows:
	@mkdir -p $(DIST_DIR)/windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/windows/doomerang.exe .

# build-linux:
# 	@mkdir -p $(DIST_DIR)/linux
# 	docker-compose run --rm build-linux

build-web:
	@mkdir -p $(DIST_DIR)/web
	GOOS=js GOARCH=wasm go build -o $(DIST_DIR)/web/doomerang.wasm .
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST_DIR)/web/
	cp assets/web/index.html $(DIST_DIR)/web/

# Build all platforms
build-all: build-mac build-windows build-web
	@echo "Built for: mac, windows, web"

# Clean dist
clean-dist:
	rm -rf $(DIST_DIR)

# itch.io deployment (requires butler installed and logged in)
deploy-mac:
	butler push $(DIST_DIR)/mac $(ITCH_USER)/$(ITCH_GAME):mac

deploy-windows:
	butler push $(DIST_DIR)/windows $(ITCH_USER)/$(ITCH_GAME):windows

deploy-linux:
	butler push $(DIST_DIR)/linux $(ITCH_USER)/$(ITCH_GAME):linux

deploy-web:
	butler push $(DIST_DIR)/web $(ITCH_USER)/$(ITCH_GAME):web

deploy-all: deploy-mac deploy-windows deploy-web
	@echo "Deployed to itch.io: mac, windows, linux, web"