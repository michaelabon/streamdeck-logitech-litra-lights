# Not a UUID, I know. Blame Elgato.
UUID := "ca.michaelabon.logitechlitra"

GO := "go"
GOFLAGS := ""
PLUGIN := UUID + ".sdPlugin"
DISTRIBUTION_TOOL := "$HOME/.bin/DistributionTool"
TARGET := "build/streamdeck-logitech-litra"


## BUILD


[macos]
build:
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ TARGET }}.exe .
    CGO_ENABLED=1 GOOS=darwin  GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ TARGET }}     .

# WSL support
[linux]
build:
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ TARGET }}.exe .
    touch {{ PLUGIN }}/{{ TARGET }} # Stream Deck complains about a missing Mac binary while on Windows. (Why??)

clean:
    rm {{ PLUGIN }}/{{ TARGET }}
    rm {{ PLUGIN }}/{{ TARGET }}.exe
    rm {{ PLUGIN }}/logs/*


## INSTALL DEV DEPENDENCIES


[windows, macos]
install: _install-submodules _install-go-tools

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
    cd go && gci write .
    gofumpt -w ./go
    golines -w ./go

[macos, linux]
lint:
    cd go && gci write .
    gofumpt -w ./go
    golines -w ./go
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


## Package the plugin for distribution to Elgato
package:
    mkdir -p build
    {{ DISTRIBUTION_TOOL }} -b -i {{ PLUGIN }} -o build/


## LOGS


[macos]
logs-streamdeck:
  cd "$HOME/Library/Logs/ElgatoStreamDeck" && cat $(ls -ltr | awk '{print $9}')

[windows]
logs-streamdeck:
  cd "%appdata%\Elgato\StreamDeck\logs\"
