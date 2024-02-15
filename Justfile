# Not a UUID, I know. Blame Elgato.
UUID := "ca.michaelabon.logitechlitra"

GO := "go"
GOFLAGS := ""
PLUGIN := UUID + ".sdPlugin"
DISTRIBUTION_TOOL := "$HOME/.bin/DistributionTool"
TARGET := "streamdeck-logitech-litra"

build:
    {{ GO }} build {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ TARGET }} -C ./go .

[macos]
link:
    ln -s \
        "{{ justfile_directory() }}/{{ PLUGIN }}" \
        "$HOME/Library/Application Support/com.elgato.StreamDeck/Plugins"

[windows]
link:
    mklink /D "%AppData%\Elgato\StreamDeck\Plugins\{{ PLUGIN }}" "{{ justfile_directory() }}/{{ PLUGIN }}"

install:
    git submodule update --init --recursive
    cd ./go && go mod tidy
    go install mvdan.cc/gofumpt@latest
    go install github.com/segmentio/golines@latest
    npm install -g @elgato/cli

lint:
    gofumpt -w ./go
    golines -w ./go

test:
    go test -C go ./...

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
