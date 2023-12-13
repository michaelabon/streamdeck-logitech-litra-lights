GO = go
GOFLAGS =
INSTALLDIR = "$(HOME)/Library/Application Support/com.elgato.StreamDeck/Plugins/dev.samwho.streamdeck.counter.sdPlugin"

TARGET = streamdeck-logitech-litra

.PHONY: test install build logs

build: $(TARGET)

$(TARGET): main.go
	$(GO) build $(GOFLAGS) -o $(TARGET) main.go

test:
	$(GO) run $(GOFLAGS) main.go -port 12345 -pluginUUID 213 -registerEvent test -info "{\"application\":{\"language\":\"en\",\"platform\":\"mac\",\"version\":\"4.1.0\"},\"plugin\":{\"version\":\"1.1\"},\"devicePixelRatio\":2,\"devices\":[{\"id\":\"55F16B35884A859CCE4FFA1FC8D3DE5B\",\"name\":\"Device Name\",\"size\":{\"columns\":5,\"rows\":3},\"type\":0},{\"id\":\"B8F04425B95855CF417199BCB97CD2BB\",\"name\":\"Another Device\",\"size\":{\"columns\":3,\"rows\":2},\"type\":1}]}"

install: build
	rm -rf $(INSTALLDIR)
	mkdir $(INSTALLDIR)
	mkdir $(INSTALLDIR)/logs
	cp manifest.json $(INSTALLDIR)
	cp -R propertyInspector $(INSTALLDIR)
	cp -R icons $(INSTALLDIR)
	cp streamdeck-logitech-litra $(INSTALLDIR)

logs:
	tail -f "logs"/streamdeck-counter.log*
