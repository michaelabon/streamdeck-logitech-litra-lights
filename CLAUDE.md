# CLAUDE.md – streamdeck-logitech-litra-lights

## [H-FIX] HID-Wildcard nie verwenden
```
hid.Enumerate(VID, 0, ...) VERBOTEN — trifft Maus/Tastatur → privilege violation (0xE00002C1)
Immer explizit: PID_LITRA_GLOW (0xC900), PID_LITRA_BEAM (0xC901), PID_LITRA_BEAM_ALT (0xC8F1)
```

## [H-FIX] Litra Beam PID
```
PID des physischen Geräts: 0xC901  (nicht 0xC8F1!)
Symptom wenn falsch: kein Fehler im Log, Licht reagiert einfach nicht
```

## Nach Code-Änderungen
```
just build → Stream Deck App Quit + neu öffnen
Logs: ca.michaelabon.logitech-litra-lights.sdPlugin/logs/ (neueste Datei)
```

## Symlink
```
Liegt in: ~/Library/Application Support/com.elgato.StreamDeck/Plugins/
Nach Projektverschiebung: just unlink && just link
```
