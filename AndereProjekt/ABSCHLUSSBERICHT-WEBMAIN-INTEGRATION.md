# Abschlussbericht: Panel-Funktionalität in WebMain

## Umgesetzte Funktionen

Die folgenden Pterodactyl-Panel-Funktionen wurden in `WebMain` integriert und an die Go/Vue-Architektur angepasst. Das bestehende Design, Branding und die Navigation von `WebMain` bleiben unverändert.

- **Wings-Kommunikation:** HTTPS für externe Nodes, HTTP nur für Loopback, keine Redirects, begrenzte Antworten, validierte UUIDs/Aktionen. Status, Suspendierung, CPU, Speicher, Datenträger, Laufzeit, Netzwerk und Logs.
- **Node-Verwaltung (Admin):** Registrieren, Verbindung testen (Systemdaten: OS, Architektur, CPU-Anzahl, Wings-Version), Token rotieren, Remote-Credentials speichern, Löschen. Tokens werden verschlüsselt gespeichert; für Rückrufe wird nur die Token-ID hinterlegt.
- **Server-Registrierung und Verwaltung (Admin):** Registrierung vorhandener Wings-Server, Umbenennung der WebMain-Anzeige, Server-Filter nach Name/Node/UUID, Zugriffsvergabe/-entzug mit granularen Rechten (Power, Konsole, Dateien, Vorschau, Download, Verwaltung, Bearbeitung, Löschung, Upload, Backups, Backup-Download), standardmäßig deaktiviert.
- **Serverdetails (Nutzer):** Zustand, Suspendierung, CPU, Speicher, Datenträger, Laufzeit, Netzwerk; cachefreie Aktualisierung alle 5 Sekunden; Ressourcengrößen in IEC-Einheiten (KiB, MiB, GiB).
- **Live-Konsole:** WebSocket-Proxy mit signiertem, kurzlebigem, schreibgeschütztem Wings-Token, Same-Origin-Prüfung, maximal drei gleichzeitige Verbindungen je Nutzer/Server; REST-Polling als Rückfall; Befehle nur für berechtigte Nutzer/Admins.
- **Power-Aktionen:** Start, Stop, Restart, Force-Stop (Force-Stop erfordert serverseitig geprüfte Bestätigung).
- **Dateiverwaltung:** Browser mit Suche nach Name/Typ, Vorschau, Download, Upload, Edit, Verzeichnis anlegen, Rename, Delete, Copy, Compress/Decompress, Chmod; getrennte Rechte.
- **Backups:** Erstellen, Wiederherstellen, Löschen (mit Bestätigung), Download via kurzlebige Einmal-Tokens als Stream, Status-Polling, Filterung nach Name/Status, **Sperren/Entsperren** (gesperrte Backups können nicht gelöscht werden).
- **Server-Lifecycle (Admin):** Install, Reinstall, Sync-Konfiguration, Delete auf Wings (jeweils mit Bestätigung; Delete entfernt anschließend auch die WebMain-Registrierung).
- **Aktivitätsverlauf:** Admin- und Nutzeransicht der neuesten Wings-Aktivitäten (Zeit, Akteur, Ereignis, Detail) für Node- und Serveraktionen. Konsolentexte und Tokens werden nicht protokolliert.
- **Lokale Wings-Backups** inklusive Download-Stream über WebMain.

## Nicht integrierte Funktionen und Begründung

- **Provisionierung neuer Server:** Wings' `POST /api/servers` ruft die Serverkonfiguration über die Pterodactyl-Panel-Remote-API ab. WebMain stellt nur Backup-Callbacks unter `/api/remote/backups/...` bereit. Benötigt: kompatible Remote-Endpunkte in WebMain.
- **Technische Startkonfiguration (Umgebungen, Startbefehle, Docker-Image):** Wings liefert `configuration.environment` inklusive Werten, die Geheimnisse enthalten können; `/api/update` ändert nur die Wings-Daemon-Konfiguration, nicht die einzelne Serverumgebung. Benötigt: geeignete Panel-API mit expliziter Geheimnisredaktion.
- **S3-Backups:** Der Wings-S3-Adapter benötigt panel-seitig signierte Upload-URLs (`GetBackupRemoteUploadURLs`). WebMain implementiert diese nicht.
- **Datenbanken, Zeitpläne, Netzwerkzuweisungen, Allocations, Eggs, Mounts, Server-Transfers, Suspension, Rebuild, Toggle-Install, Force-Destroy:** Werden in Pterodactyl durch die Panel-API verwaltet; die in WebMain integrierte Wings-API stellt keine passenden Endpunkte bereit. Direkte Datenbank-/Container-Manipulation ohne unterstützte API wird nicht als Ersatz implementiert.
- **Benachrichtigungen (E-Mail/Push), SSH-Keys, panel-seitige Administration/Systemeinstellungen:** Keine Wings-Entsprechung; außerhalb des Wings-basierten Integrationsumfangs.

## Wichtige geänderte Dateien und Strukturen

- `WebMain/services/wings/client.go` — Wings-HTTP-Client (inkl. Lifecycle-Methoden).
- `WebMain/models/wings/` — `node.go`, `server.go`, `access.go`, `backup.go`, `activity.go` (neu).
- `WebMain/models/novroforge_migrations/` — Migrationen `v18b`–`v18p` (Nodes, Server, Access, Dateirechte, Remote-Credentials, Backups, Backup-Download, Aktivität, Backup-Lock).
- `WebMain/routers/web/admin/wings.go`, `WebMain/routers/web/user/wings.go`, `WebMain/routers/web/user/wings_backups.go` — Handler.
- `WebMain/routers/web/web.go` — Routenregistrierung.
- `WebMain/templates/admin/wings_nodes.tmpl`, `WebMain/templates/wings/server.tmpl`, `WebMain/templates/wings/backups.tmpl`, `WebMain/templates/wings/files.tmpl` — UI.
- `WebMain/options/locale_next/locale_en-US.json` — Locale-Schlüssel (1495 Schlüssel).
- `WebMain/vendor/github.com/gorilla/websocket/` — vendoredes Modul.

## Ausgeführte Prüfungen

- `go test -count=1 ./services/wings/` — bestanden.
- `go test -c ./models/wings/` (Linux-Ziel) — Test-Binaries kompilieren.
- `go vet ./services/wings/ ./models/wings/ ./routers/web/admin/ ./routers/web/user/` — bestanden.
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` — bestanden.
- `gofmt -l` auf geänderten Go-Dateien — keine Abweichungen.
- Locale-JSON-Parsing — valide, keine Duplikate.
- Markdownlint mit `WebMain/.markdownlint.yaml` — bestanden.
- `git diff --check` — keine Whitespace-Fehler.
- ESLint und Webpack-Produktionsbuild für Frontend-Änderungen — bestanden (frühere Schritte).

Hinweis: Ein nativer Windows-Build ist für dieses Linux-zentrierte Projekt nicht möglich (Unix-spezifische `syscall`-/`unix`-Nutzung in unverwandten Modulen). Tests/Builds laufen über die Linux-Cross-Toolchain (`C:\Users\fates\AppData\Local\Programs\GoToolchain\go\bin`).

## Manuelle Konfigurations-/Betriebsschritte

1. Go-Toolchain (`go1.27.3`) liegt nicht auf `PATH`; für Builds/Tests `GOTOOLCHAIN=local` und den Pfad `C:\Users\fates\AppData\Local\Programs\GoToolchain\go\bin` verwenden.
2. Datenbankmigrationen (`v18b`–`v18p`) laufen beim Start automatisch; sicherstellen, dass die WebMain-Datenbank erreichbar ist.
3. Wings-Nodes müssen mit HTTPS (oder Loopback-HTTP) registriert werden; Node-Bearer-Tokens werden verschlüsselt gespeichert.
4. Für lokale Backups müssen Remote-Callback-Credentials (`/admin/wings/nodes/{id}/remote-credentials`) gesetzt sein, damit Wings Status-Callbacks an WebMain senden kann.
5. Keine echten Tokens in Quellcode oder Dokumentation hinterlegen; Node-Tokens werden erst zur Laufzeit eingegeben.
6. Die verbleibenden panel-verwalteten Bereiche (Datenbanken, Zeitpläne, Allokationen, S3-Backups, Provisionierung) erfordern eine zukünftige Panel-Remote-API in WebMain.
