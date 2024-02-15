GO := "go"
GOFLAGS := ""
PLUGIN := "ca.michaelabon.logitechlitra.sdPlugin"
DISTRIBUTION_TOOL := "$HOME/.bin/DistributionTool"
OUTPUT := "streamdeck-logitech-litra"

build: streamdeck-logitech-litra

[macos]
streamdeck-logitech-litra:
    GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}.exe .
    GOOS=darwin  GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}     .

# WSL support
[linux]
streamdeck-logitech-litra:
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 {{ GO }} build -C go {{ GOFLAGS }} -o ../{{ PLUGIN }}/{{ OUTPUT }}.exe .
    touch {{ PLUGIN }}/{{ OUTPUT }} # Stream Deck complains about a missing Mac binary while on Windows. (Why??)


[windows]
install: install-go-tools

[macos]
install: install-go-tools


## WSL support
[linux]
install: install-go-tools
    sudo apt install gcc-mingw-w64

install-go-tools:
    go install github.com/daixiang0/gci@latest
    go install mvdan.cc/gofumpt@latest



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
    cd go && gci write .
    cd go && gofumpt -l -w .


## Package the plugin for distribution to Elgato
package:
    mkdir build
    {{ DISTRIBUTION_TOOL }} -b -i {{ PLUGIN }} -o build/
