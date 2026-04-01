# Übergabe: streamdeck-logitech-litra-lights

## Letzter Stand (2026-04-01)
Plugin funktioniert wieder. Zwei Bugs gefixt:
1. Symlink zeigte auf alten Pfad (`~/streamdeck-logitech-litra-lights/`) → auf `~/Claude-Projekte/...` korrigiert
2. `hid.Enumerate(VID, 0, ...)` traf Maus/Tastatur → privilege violation; Litra Beam PID war als `0xC8F1` hinterlegt, tatsächlich `0xC901`

## Nächster Schritt
Kein offener Task. Optional: Version im manifest.json auf 1.0.7.0 bumpen (`just bump 1.0.7.0`).

## Offene Punkte
- Plugin noch nicht im Elgato Marketplace eingereicht
- Kein x86_64 macOS Build (nur arm64) — kein Problem für aktuelles Gerät (Apple Silicon)
