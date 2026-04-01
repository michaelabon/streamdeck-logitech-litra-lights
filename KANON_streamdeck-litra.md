# KANON: streamdeck-logitech-litra-lights
// Wissensdatenbank für das Stream Deck Logitech Litra Plugin

---

## Projekt-Übersicht

Stream Deck Plugin (Go) zur Steuerung von Logitech Litra Lampen via USB-HID.

- **Repo:** https://github.com/serieger21/streamdeck-logitech-litra-lights
- **Lokal:** `~/Claude-Projekte/streamdeck-logitech-litra-lights/`
- **Plugin-Install:** Symlink in `~/Library/Application Support/com.elgato.StreamDeck/Plugins/`
- **Build-Tool:** `just` (justfile im Root)
- **Sprache:** Go 1.24+ mit CGO (hidapi)

---

## Architektur

```
go/
  main.go              – Plugin-Einstiegspunkt, Stream Deck Event-Handler
  drawing.go           – 72×72px Farbgradient-Bild für Button (Kelvin + Brightness)
  internal/logitech_hid/
    bytes.go           – HID-Protokoll: 20-Byte-Kommandos für Brightness/Temperature/On/Off
    bytes_test.go      – Unit-Tests

ca.michaelabon.logitech-litra-lights.sdPlugin/
  manifest.json        – Plugin-Metadaten, Versions- und Binary-Pfade
  propertyInspector/   – HTML/JS UI für Einstellungen im Stream Deck
  build/               – Kompilierte Binaries (arm64 macOS, .exe Windows)
  logs/                – Logfiles des laufenden Plugins
  icons/               – Plugin-Icons
```

---

## HID-Protokoll (bytes.go)

Logitech erwartet immer **20 Bytes** pro Kommando:

| Kommando        | Bytes (hex)                          |
|-----------------|--------------------------------------|
| Licht an        | `11 FF 04 1C 01 00 00 ...`           |
| Licht aus       | `11 FF 04 1C 00 00 00 ...`           |
| Brightness      | `11 FF 04 4C 00 <wert> 00 ...`       |
| Temperature     | `11 FF 04 9C <hi> <lo> 00 ...`       |

- **Brightness:** Byte-Wert 0x14–0xFA (20–250), mapped von 1–100%
- **Temperature:** Big-Endian uint16, Bereich 2700–6500 K

---

## Bekannte Logitech PIDs

| Gerät           | VID    | PID    |
|-----------------|--------|--------|
| Litra Glow      | 0x046d | 0xC900 |
| Litra Beam      | 0x046d | 0xC901 |
| Litra Beam (alt)| 0x046d | 0xC8F1 |

**[H-FIX 2026-04-01]** PID 0xC8F1 war ursprünglich als einziger Beam-PID eingetragen — das ist falsch.
Der physisch vorhandene Litra Beam hat PID 0xC901. Beide PIDs müssen enumeriert werden.
Symptom des falschen PIDs: kein Fehler im Log, Licht reagiert einfach nicht.

---

## macOS HID-Permission-Falle

**[H-FIX 2026-04-01]** `hid.Enumerate(VID, 0, ...)` (PID=0 = Wildcard) enumeriert **alle** Logitech-Geräte —
also auch MX Keys (0xB015) und MX Master Maus (0xB33B). Diese sind durch macOS Input-Monitoring geschützt
und erzeugen:
```
privilege violation (0xE00002C1) (iokit/common)
```
**Regel:** Immer explizit nach bekannten Litra-PIDs enumerieren, niemals PID=0 verwenden.

---

## Build-Workflow

```bash
# Einmalig (Abhängigkeiten + Symlink):
just install build link

# Nach Code-Änderungen:
just build
# → Stream Deck App neu starten (Quit + öffnen)

# Tests:
just test

# Logs lesen:
ls -lt ca.michaelabon.logitech-litra-lights.sdPlugin/logs/ | head -3
# → neueste Datei öffnen
```

---

## Symlink

Der Stream Deck Plugin-Ordner ist per Symlink verknüpft:
```
~/Library/Application Support/com.elgato.StreamDeck/Plugins/ca.michaelabon.logitech-litra-lights.sdPlugin
→ ~/Claude-Projekte/streamdeck-logitech-litra-lights/ca.michaelabon.logitech-litra-lights.sdPlugin
```
Bei Verschieben des Projektverzeichnisses: `just unlink && just link` aus dem neuen Verzeichnis.

---

## Property Inspector (UI)

- HTML/JS unter `ca.michaelabon.logitech-litra-lights.sdPlugin/propertyInspector/pi.html`
- Slider für Temperature (2700–6500 K) und Brightness (1–100%)
- Slider-Farbe reagiert dynamisch auf Kelvin-Wert (RGB-Berechnung per Näherungsformel)

---

## Versionshistorie

| Version | Datum      | Änderung |
|---------|------------|----------|
| 1.0.6.0 | 2026-01-16 | Build mit Litra Beam alt-PID 0xC8F1 |
| 1.0.7.0 | 2026-04-01 | PID 0xC901 ergänzt, HID-Wildcard-Bug gefixt, Symlink auf neuen Pfad |
