# Not a UUID, I know. Blame Elgato.
UUID := "ca.michaelabon.logitechlitra"

GO := "go"
GOFLAGS := ""
PLUGIN := UUID + ".sdPlugin"
DISTRIBUTION_TOOL := "$HOME/.bin/DistributionTool"
OUTPUT := "streamdeck-logitech-litra"

[macos]
build:
    GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}.exe .
    GOOS=darwin  GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}     .

# WSL support
[linux]
build:
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}.exe .
    touch {{ PLUGIN }}/{{ OUTPUT }} # Stream Deck complains about a missing Mac binary while on Windows. (Why??)


## WSL support
[linux]
install: install-base
    sudo apt install gcc-mingw-w64

[macos]
install: install-base

install-base:
    git submodule update --init --recursive
    cd ./go && go mod tidy
    go install mvdan.cc/gofumpt@latest
    go install github.com/segmentio/golines@latest
    npm install -g @elgato/cli



[macos]
link:
    ln -s \
        "{{ justfile_directory() }}/{{ PLUGIN }}" \
        "$HOME/Library/Application Support/com.elgato.StreamDeck/Plugins"

[windows]
link:
    mklink /D "%AppData%\Elgato\StreamDeck\Plugins\{{ PLUGIN }}" "{{ justfile_directory() }}/{{ PLUGIN }}"

test:
    go test -C go ./...


lint:
    gofumpt -w ./go
    golines -w ./go

[macos]
debug:
    open "http://localhost:23654/"

[windows]
debug:
    start "" "http://localhost:23654/"

start:
    streamdeck restart {{ UUID }}
restart: start

## Package the plugin for distribution to Elgato
package:
    mkdir build
    {{ DISTRIBUTION_TOOL }} -b -i {{ PLUGIN }} -o build/
