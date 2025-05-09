# Not a UUID, I know. Blame Elgato.
UUID := "ca.michaelabon.logitech-litra-lights"

GO := "go"
GOFLAGS := ""
PLUGIN := UUID + ".sdPlugin"
TARGET := "build/streamdeck-logitech-litra-lights"
HIDAPI_DIR := "hidapi"

set windows-shell := ["powershell.exe", "-c"]

## BUILD


[macos]
build: setup-hidapi
    # Windows build
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
    CGO_LDFLAGS="-static-libgcc -L$(pwd)/{{ HIDAPI_DIR }}/windows -lhidapi" \
    {{ GO }} build \
    -C go \
    {{ GOFLAGS }} \
    -o ../{{ PLUGIN }}/{{ TARGET }}.exe \
    .

    # Copy Windows DLL as fallback in case static linking fails
    cp $(pwd)/{{ HIDAPI_DIR }}/windows/hidapi.dll {{ PLUGIN }}/

    # macOS build - build for both architectures
    # Build for arm64 (Apple Silicon)
    CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
    CGO_LDFLAGS="-L$(pwd)/{{ HIDAPI_DIR }}/macos -lhidapi" \
    {{ GO }} build \
    -C go \
    {{ GOFLAGS }} \
    -o ../{{ PLUGIN }}/{{ TARGET }}_arm64 \
    .

    # Copy macOS dylib (with force flag to overwrite if needed)
    rm -f {{ PLUGIN }}/libhidapi.dylib || true
    cp -f $(pwd)/{{ HIDAPI_DIR }}/macos/libhidapi.dylib {{ PLUGIN }}/

    # Fix macOS binary to reference the dylib relatively
    install_name_tool -change libhidapi.dylib @executable_path/libhidapi.dylib {{ PLUGIN }}/{{ TARGET }}_arm64

# WSL support
[linux]
build: setup-hidapi
    # Windows build - make sure we use the correct path
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
    CGO_LDFLAGS="-static-libgcc -L$(pwd)/{{ HIDAPI_DIR }}/windows -lhidapi" \
    {{ GO }} build \
    -C go \
    {{ GOFLAGS }} \
    -o ../{{ PLUGIN }}/{{ TARGET }}.exe\
    .

    # Copy Windows DLL as fallback in case static linking fails
    cp $(pwd)/{{ HIDAPI_DIR }}/windows/hidapi.dll {{ PLUGIN }}/

    # Create dummy macOS file
    touch {{ PLUGIN }}/{{ TARGET }} # Stream Deck complains about a missing Mac binary while on Windows. (Why??)

# Setup HIDAPI library directory with required files
setup-hidapi:
    mkdir -p {{ HIDAPI_DIR }}/windows {{ HIDAPI_DIR }}/macos

    # Download and extract Windows hidapi.dll (x64 version)
    if [ ! -f {{ HIDAPI_DIR }}/windows/hidapi.dll ]; then \
      curl -L https://github.com/libusb/hidapi/releases/download/hidapi-0.14.0/hidapi-win.zip -o {{ HIDAPI_DIR }}/hidapi-win.zip && \
      unzip -j {{ HIDAPI_DIR }}/hidapi-win.zip 'x64/hidapi.dll' -d {{ HIDAPI_DIR }}/windows && \
      rm -f {{ HIDAPI_DIR }}/hidapi-win.zip; \
    fi

    # For macOS, install hidapi via homebrew and copy the dylib
    if [ ! -f {{ HIDAPI_DIR }}/macos/libhidapi.dylib ]; then \
      brew list hidapi || brew install hidapi && \
      cp $(brew --prefix hidapi)/lib/libhidapi.dylib {{ HIDAPI_DIR }}/macos/; \
    fi

clean:
    rm -f {{ PLUGIN }}/{{ TARGET }}
    rm -f {{ PLUGIN }}/{{ TARGET }}_arm64
    rm -f {{ PLUGIN }}/{{ TARGET }}.exe
    rm -f {{ PLUGIN }}/hidapi.dll
    rm -f {{ PLUGIN }}/libhidapi.dylib
    rm -f {{ PLUGIN }}/logs/*
    rm -rf ./hidapi

## INSTALL DEV DEPENDENCIES
[windows]
install: _install-submodules _install-go-tools

[macos]
install: _install-submodules _install-go-tools
    brew install mingw-w64

[linux] ## WSL support
install: _install-go-tools
    sudo apt install gcc-mingw-w64

_install-submodules:
    git submodule update --init --recursive
    cd ./go && go mod download

_install-go-tools:
    go install github.com/daixiang0/gci@latest
    go install mvdan.cc/gofumpt@latest
    go install github.com/segmentio/golines@latest


## LINK
## You only need to do this one time.
## It connects your output directory to your Stream Deck


[macos]
link:
    ln -s \
        "{{ justfile_directory() }}/{{ PLUGIN }}" \
        "$HOME/Library/Application Support/com.elgato.StreamDeck/Plugins"

[windows]
link:
    mklink /D "%AppData%\Elgato\StreamDeck\Plugins\{{ PLUGIN }}" "{{ justfile_directory() }}/{{ PLUGIN }}"

[macos]
unlink:
    unlink "$HOME/Library/Application Support/com.elgato.StreamDeck/Plugins/{{ PLUGIN }}"


## TEST
## Run unit tests


test:
    go test -C go ./...


## LINT
## Ensure that all the files are formatted correctly.


[windows]
lint:
    cd go && golangci-lint run

[macos, linux]
lint:
    cd go && golangci-lint run
    find ./go ./{{ PLUGIN }}/icons -type f -name '*.svg' -exec xmllint --pretty 2 --output '{}' '{}' \;


## DEBUG & RESTART
## Useful for local development

[macos]
debug:
    open "http://localhost:23654/"

[windows]
debug:
    start "" "http://localhost:23654/"

start:
    npx streamdeck restart {{ UUID }}
restart: start


## PACKAGE


validate:
    cd "{{ PLUGIN }}" && npx streamdeck validate

## Package the plugin for distribution to Elgato
package: build
    mkdir -p build
    npx streamdeck pack -o build/ {{ PLUGIN }}


## LOGS


[macos]
logs-streamdeck:
  cd "$HOME/Library/Logs/ElgatoStreamDeck" && cat $(ls -ltr | awk '{print $9}')

[windows]
logs-streamdeck:
  cd "%appdata%\Elgato\StreamDeck\logs\"


## VERSIONING


bump version:
  yq --inplace --prettyPrint --output-format json '.Version = "{{ version }}"' {{ PLUGIN }}/manifest.json
