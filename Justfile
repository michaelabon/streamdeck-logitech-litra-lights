GO := "go"
GOFLAGS := ""
PLUGIN := "ca.michaelabon.logitechlitra.sdPlugin"
DISTRIBUTION_TOOL := "$HOME/.bin/DistributionTool"

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

## Package the plugin for distribution to Elgato
package:
    mkdir build
    {{ DISTRIBUTION_TOOL }} -b -i {{ PLUGIN }} -o build/
