## Goal
Set up build targets for multiple platforms (Windows, Mac, Linux, Web/WASM) and automate deployment to itch.io using butler.

## Design

### Target Platforms
| Platform | Cross-compile from Mac? | Output |
|----------|------------------------|--------|
| macOS (arm64) | ✅ Native | `doomerang-mac` |
| macOS (amd64) | ✅ Native | `doomerang-mac-intel` |
| Windows | ✅ Yes | `doomerang.exe` |
| Linux | ❌ Needs Linux runner* | `doomerang-linux` |
| Web (WASM) | ✅ Yes | `doomerang.wasm` + HTML |

*Linux builds require Cgo and can't be cross-compiled from Mac. We will use docker and docker-compose to build from linux using a prebuilt golang image with the debian version "trixie" from https://hub.docker.com/_/golang/ 

### Build Output Structure
```
dist/
├── mac/
│   └── doomerang
├── mac-intel/
│   └── doomerang
├── windows/
│   └── doomerang.exe
├── linux/
│   └── doomerang
└── web/
    ├── doomerang.wasm
    ├── wasm_exec.js
    └── index.html
```

### itch.io Deployment
- **Tool:** [butler](https://itch.io/docs/butler/) (itch.io's CLI uploader)
- **Channel naming:** `mac`, `windows`, `linux`, `web` (kebab-case)
- **Delta uploads:** butler only uploads changed bytes after first push

## Build Commands

### Native Mac Build
```bash
go build -o dist/mac/doomerang .
```

### Windows Cross-Compile (from Mac)
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/windows/doomerang.exe .
```

### Web/WASM Build
```bash
GOOS=js GOARCH=wasm go build -o dist/web/doomerang.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/web/
# Copy index.html template
```

### Linux Build (via Docker)

Uses `golang:trixie` image from Docker Hub.

**docker-compose.yml:**
```yaml
services:
  build-linux:
    image: golang:trixie
    working_dir: /app
    volumes:
      - .:/app
      - go-cache:/go/pkg/mod
    command: go build -o dist/linux/doomerang .
    environment:
      - GOOS=linux
      - GOARCH=amd64

volumes:
  go-cache:
```

**Build command:**
```bash
docker-compose run --rm build-linux
```

## Makefile Targets

Add these targets to the existing Makefile:

```makefile
# Build directories
DIST_DIR := dist

# Platform builds
build-mac:
	@mkdir -p $(DIST_DIR)/mac
	go build -o $(DIST_DIR)/mac/doomerang .

build-mac-intel:
	@mkdir -p $(DIST_DIR)/mac-intel
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/mac-intel/doomerang .

build-windows:
	@mkdir -p $(DIST_DIR)/windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/windows/doomerang.exe .

build-linux:
	@mkdir -p $(DIST_DIR)/linux
	docker-compose run --rm build-linux

build-web:
	@mkdir -p $(DIST_DIR)/web
	GOOS=js GOARCH=wasm go build -o $(DIST_DIR)/web/doomerang.wasm .
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST_DIR)/web/
	cp assets/web/index.html $(DIST_DIR)/web/

# Build all platforms
build-all: build-mac build-mac-intel build-windows build-linux build-web
	@echo "Built for: mac, mac-intel, windows, linux, web"

# Clean dist
clean-dist:
	rm -rf $(DIST_DIR)

# itch.io deployment (requires butler installed and logged in)
ITCH_USER := your-username
ITCH_GAME := doomerang

deploy-mac:
	butler push $(DIST_DIR)/mac $(ITCH_USER)/$(ITCH_GAME):mac

deploy-mac-intel:
	butler push $(DIST_DIR)/mac-intel $(ITCH_USER)/$(ITCH_GAME):mac-intel

deploy-windows:
	butler push $(DIST_DIR)/windows $(ITCH_USER)/$(ITCH_GAME):windows

deploy-linux:
	butler push $(DIST_DIR)/linux $(ITCH_USER)/$(ITCH_GAME):linux

deploy-web:
	butler push $(DIST_DIR)/web $(ITCH_USER)/$(ITCH_GAME):web

deploy-all: deploy-mac deploy-mac-intel deploy-windows deploy-linux deploy-web
	@echo "Deployed to itch.io: mac, mac-intel, windows, linux, web"
```

## Web Build HTML Template

Create `assets/web/index.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Doomerang</title>
    <style>
        body { margin: 0; padding: 0; background: #000; }
        canvas { display: block; margin: 0 auto; }
    </style>
</head>
<body>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("doomerang.wasm"), go.importObject).then(result => {
            go.run(result.instance);
        });
    </script>
</body>
</html>
```

## Implementation Tasks

### Setup
- [ ] Install butler: `brew install butler` or download from itch.io
- [ ] Login to butler: `butler login`
- [ ] Create itch.io game page (if not exists)
- [ ] Update `ITCH_USER` and `ITCH_GAME` in Makefile

### Makefile Targets
- [ ] Add `build-mac` target
- [ ] Add `build-mac-intel` target
- [ ] Add `build-windows` target
- [ ] Add `build-web` target with WASM + HTML
- [ ] Add `build-all` convenience target
- [ ] Add `clean-dist` target

### Web Build
- [ ] Create `assets/web/index.html` template
- [ ] Test WASM build locally with `wasmserve`
- [ ] Verify audio works in browser (may need user interaction)

### Deployment
- [ ] Add `deploy-mac` target
- [ ] Add `deploy-windows` target
- [ ] Add `deploy-web` target
- [ ] Add `deploy-all` convenience target
- [ ] Test deployment to itch.io

### Linux Build (Docker)
- [ ] Create `docker-compose.yml` with build-linux service
- [ ] Add `build-linux` Makefile target using docker-compose
- [ ] Test Linux build via Docker

## Files to Create/Modify
- `Makefile` - Add build and deploy targets
- `docker-compose.yml` - Linux build container config
- `assets/web/index.html` - WASM HTML template

## Dependencies
- **Go 1.21+** - For WASM support
- **Docker** - For Linux builds
- **butler** - itch.io CLI tool ([install docs](https://itch.io/docs/butler/))

## Testing

### Test WASM locally
```bash
go run github.com/hajimehoshi/wasmserve@latest .
# Open http://localhost:8080
```

### Test butler (dry run)
```bash
butler push dist/web your-username/doomerang:web --dry-run
```

## References
- [Ebitengine WebAssembly Docs](https://ebitengine.org/en/documents/webassembly.html)
- [Butler Push Documentation](https://itch.io/docs/butler/pushing.html)
- [Ebitengine Cross-Compilation FAQ](https://ebitengine.org/en/documents/faq.html)
- See `itch-io-uploads.md` for step-by-step itch.io project setup guide
