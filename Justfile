GO := "go"
GOFLAGS := ""
PLUGIN := "ca.michaelabon.logitechlitra.sdPlugin"

build: streamdeck-logitech-litra

streamdeck-logitech-litra:
    {{ GO }} build {{ GOFLAGS }} -o ../{{ PLUGIN }}/streamdeck-logitech-litra -C go .

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

# From https://github.com/bobheadxi/readable
lint:
    readable fmt README.md
