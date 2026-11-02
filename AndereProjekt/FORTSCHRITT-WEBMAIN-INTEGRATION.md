# Fortschritt: WebMain-Integration

Diese Datei dokumentiert den Fortschritt bei der Integration der Panel-Funktionen aus dem Referenzprojekt in `WebMain`.

## Projektpfade

- **Referenzprojekt:** `C:\Users\fates\Documents\GitHub\novroforge\AndereProjekt\panel-release-v2.0.0-beta.2`
- **Zielprojekt:** `C:\Users\fates\Documents\GitHub\novroforge\WebMain`

## Vorgaben

- Das Referenzprojekt bleibt unverändert.
- Änderungen werden ausschließlich in `WebMain` vorgenommen.
- Das vorhandene Design, Branding und die Benutzerführung von `WebMain` bleiben erhalten.
- Funktionen gelten erst als erledigt, wenn sie implementiert und passend getestet wurden.
- Nicht umsetzbare Funktionen und Blocker werden mit Begründung dokumentiert.

## Statusübersicht

Status: **In Arbeit**

### Gesamtfortschritt: ca. 30 % (Schätzung)

`[###.......] 30 %`
- Die Schätzung bezieht sich auf den geplanten Gesamtumfang eines Panel-Integrationsprojekts (Analyse, sichere Wings-Anbindung, Nodes/Server, Benutzerzugriff und weitere Panel-Funktionen), nicht auf die Anzahl der geänderten Dateien. Die vollständige Funktionsinventur und Nutzerberechtigungen fehlen noch.

- [ ] Analyse beider Projekte abgeschlossen
- [ ] Funktionen des Referenzprojekts erfasst und mit `WebMain` abgeglichen
- [ ] Integrationsumfang und Abhängigkeiten dokumentiert
- [ ] Umsetzung abgeschlossen
- [ ] Tests, Linter und Builds ausgeführt
- [ ] Abschlussbericht erstellt

Aktueller Umsetzungsschritt: sichere Wings-Node-Konfiguration, Admin-Verwaltung und erste Benutzeransicht mit expliziter Serverzuordnung. Node-Tokens werden mit WebMain-Verschlüsselung gespeichert. Zugewiesene Nutzer sehen Serverstatus und letzte Konsolenzeilen; Power-Aktionen und Konsolenbefehle erfordern jeweils gesonderte Admin-Freigaben. Live-Konsole und umfangreiche Panel-Funktionen fehlen weiterhin.

## Erste Architekturprüfung

- `WebMain` ist ein Go-Projekt (Go 1.26) mit einer bestehenden Vue-3-Weboberfläche und eigener Router-, Modell- und API-Struktur.
- Das Referenzprojekt ist Pterodactyl Panel 2.0: Backend mit PHP 8.3 und Laravel 13, Frontend mit React 19; es besitzt umfangreiche eigene APIs und Datenmodelle.
- Im Referenzprojekt sind unter anderem Client-, Application- und Admin-APIs sowie Serveraktionen für Konsole, Leistung, Dateien, Backups, Datenbanken, Zeitpläne, Nodes, Eggs, Benutzer und Berechtigungen vorhanden.
- Eine passende Pterodactyl-/Wings-Integration wurde bei der ersten Suche in `WebMain` nicht gefunden. Die vollständige Prüfung des Zielprojekts steht noch aus.
- Wegen der unterschiedlichen Architekturen ist eine vollständige Integration eine umfangreiche, schrittweise Neuumsetzung. Das Referenzprojekt darf nicht einfach in `WebMain` kopiert werden; das bestehende WebMain-Design bleibt unverändert.
- Das vorhandene `Wings`-Projekt stellt Server-Endpunkte für Status, Konsole, Power-Aktionen und weitere Ressourcen bereit. Administrative Wings-Aufrufe erfordern ein konfiguriertes Bearer-Token pro Node.
- Vor Beginn dieser Integration hatte WebMain weder eine Wings-Node-Konfiguration noch Server-Zugriffsmodelle für Panel-Benutzer. Die Node-Konfiguration, Serverregistrierung und eine erste explizite Benutzerzuordnung mit getrennten Power-/Konsolenrechten sind nun implementiert. Es werden keine echten Tokens in Quellcode oder Dokumentation eingetragen.
- Node-Verwaltung und Rechtevergabe bleiben Admin-only. Normale Nutzer sehen ausschließlich explizit zugewiesene Server.

## Funktions- und Umsetzungsübersicht

| Bereich/Funktion | Im Referenzprojekt vorhanden | Status in WebMain | Umsetzung / Datei / Hinweis |
|---|---|---|---|
| Wings-Kommunikation: Serverdetails, Logs, Power und Konsole | Ja | In Arbeit | `WebMain/services/wings/client.go`; Status, letzte 100 Logs, Power und Befehle in Admin- und berechtigungsgeprüften Nutzerseiten; Live-Streaming fehlt |
| Nodes in der Admin-Oberfläche registrieren, Verbindung testen und entfernen | Ja | In Arbeit | `WebMain/models/wings/node.go`, Migration `v18b_add-wings-nodes.go`, `/admin/wings/nodes`; Token wird verschlüsselt gespeichert |
| Serverübersicht und Serverdetails | Ja | In Arbeit | Adminübersicht und Nutzeransicht unter `/user/servers`, nur explizit zugewiesene Server; Status und Ressourcen |
| Server erstellen, bearbeiten und verwalten | Ja | Noch zu prüfen | |
| Status und Ressourcenübersicht | Ja | Noch zu prüfen | |
| Live-Konsole und Befehle | Ja | In Arbeit | Letzte 100 Konsolenzeilen; Befehle nur für explizit berechtigte Nutzer oder Administratoren; Live-/Streaming-Konsole fehlt |
| Starten, Stoppen und Neustarten | Ja | In Arbeit | Power-Aktionen nur für explizit berechtigte Nutzer oder Administratoren |
| Konfigurationen und Umgebungen | Ja | Noch zu prüfen | |
| Benutzer, Rollen und Berechtigungen | Ja | In Arbeit | Admin kann WebMain-Nutzern Serverzugriff sowie getrennte Power-/Konsolenrechte vergeben oder entziehen; vollständiges Rollenmodell fehlt |
| Aktivitätsverlauf und Benachrichtigungen | Ja | Noch zu prüfen | |
| Dateien, Backups, Datenbanken, Netzwerke und Zeitpläne | Ja | Noch zu prüfen | |
| Nodes, Allocations und Eggs | Ja | Noch zu prüfen | |
| Administration und Systemeinstellungen | Ja | Noch zu prüfen | |
| Weitere relevante Funktionen | Noch zu prüfen | Noch zu prüfen | |

Mögliche Statuswerte: **Noch zu prüfen**, **Offen**, **In Arbeit**, **Erledigt**, **Nicht umgesetzt**.

## Prüfungen

| Prüfung | Ergebnis | Hinweise |
|---|---|---|
| Tests | Nicht ausführbar | `go test ./services/wings` wurde erneut versucht; Go-Toolchain ist in der aktuellen Shell nicht installiert/verfügbar. Für das Abrufen der Konsolenzeilen wurde ein gezielter Test ergänzt, konnte aber ebenfalls nicht ausgeführt werden. |
| Linter / Formatierer | Nicht ausführbar | `gofmt` wurde versucht; Go-Toolchain ist in der aktuellen Shell nicht installiert/verfügbar. |
| Localization JSON | Bestanden | `options/locale_next/locale_en-US.json` mit PowerShell-JSON-Parser eingelesen. |
| Git whitespace check | Bestanden | `git diff --check` meldet keine Fehler. |
| Build / Typprüfung | Noch nicht ausgeführt | Nach Bereitstellung der Go-Toolchain erforderlich. |

## Blocker und nicht umgesetzte Funktionen

- **Noch offen:** Rotation der Wings-Node-Tokens sowie weitergehende Rollen-, Datei- und Ressourcenberechtigungen.
- Die aktuelle Umgebung stellt `go` und `gofmt` nicht bereit. Daher konnten die hinzugefügten Go-Änderungen hier noch nicht kompiliert oder getestet werden.
- Benutzerkonten können Servern explizit zugeordnet werden. Status/Logs sind für zugewiesene Nutzer lesbar; Power-Aktionen und Konsolenbefehle sind per separater Freigabe geschützt.

## Fortschrittsprotokoll

| Datum | Änderung / Ergebnis | Geänderte Bereiche | Nächster Schritt |
|---|---|---|---|
| 2026-11-02 | Erste Architekturprüfung: Go/Vue-Zielprojekt und Laravel/React-Referenzpanel festgestellt; Integrationsumfang ist groß und muss phasenweise umgesetzt werden. | Beide Projekte read-only geprüft; Fortschrittsdokument aktualisiert. | WebMain-Funktionen und Erweiterungspunkte vollständig inventarisieren, danach Integrationsschritte festlegen. |
| 2026-11-02 | Wings-Routen und Authentifizierung geprüft: Serveraktionen benötigen das Bearer-Token der jeweiligen Node; WebMain hat in der ersten Suche keine Panel-Integration gefunden. | Wings-Router und Pterodactyl-Daemon-Client read-only geprüft; Fortschrittsdokument aktualisiert. | Sichere Node-/Server-Datenhaltung und Berechtigungsprüfung entwerfen, dann einen getesteten Kernbereich in WebMain implementieren. |
| 2026-11-02 | Wings-HTTP-Client ergänzt: HTTPS für externe Nodes, HTTP nur für Loopback, keine Redirects, begrenzte Antworten und validierte UUIDs/Aktionen. Go-Tests und Formatierung konnten wegen fehlender Go-Toolchain nicht ausgeführt werden. | `WebMain/services/wings/client.go`, `WebMain/services/wings/client_test.go` | Go-Toolchain bereitstellen, Tests/Formatierung ausführen, anschließend verschlüsselte Node-Verwaltung und autorisierte WebMain-Endpunkte ergänzen. |
| 2026-11-02 | Admin-Verwaltung für Wings-Nodes ergänzt: Migration und Modell mit XChaCha20-Poly1305-Verschlüsselung, HTTPS-Validierung sowie Registrierung/Löschung unter `/admin/wings/nodes`. Token wird nicht in der Liste angezeigt. | Modelle, Migration, Admin-Handler/Routen, Admin-Navigation und `wings_nodes.tmpl`; Locale-JSON validiert und `git diff --check` bestanden. | Go-Toolchain bereitstellen und alle Änderungen kompilieren/testen; danach Wings-Serverzuordnung und Rechteprüfung aufbauen. |
| 2026-11-02 | Admin-Verbindungstest ergänzt: authentifizierter Aufruf von Wings `/api/system`; Fehler werden nutzerverständlich angezeigt, Token wird nicht ausgegeben. Requests und Konsolenbefehle im Client zusätzlich begrenzt. | Wings-Client, Node-Modell, Admin-Handler/Routen und Nodes-Ansicht. | Erneute Build-/Testvalidierung nach Verfügbarkeit von Go; anschließend sichere Serverzuordnung und Rechteprüfung. |
| 2026-11-02 | Benutzerzuordnung ergänzt: Admins können aktive WebMain-Konten einem Server zuweisen und Power-/Konsolenrechte separat vergeben. Nutzer sehen unter `/user/servers` nur ihre zugewiesenen Server und können je nach Freigabe Aktionen ausführen. Beim Entfernen einer Serverregistrierung werden ihre Zugriffszuordnungen in einer Transaktion gelöscht. Fortschritt auf ca. 30 % aktualisiert. | Wings-Zugriffsmodell und Migration `v18d`, Admin-/Nutzerrouten, Berechtigungsprüfungen, Benutzeransichten, Navigation und Locale. | Go-Toolchain für Build/Tests bereitstellen; danach Live-Konsole, Datei- und Backup-Funktionen sowie weitere Panel-Bereiche implementieren. |
| 2026-11-02 | Nutzer-Serverseiten in die normale Desktop- und mobile Navigation aufgenommen; Serverrouten liegen unter `/user/servers`, um Konflikte mit öffentlichen Benutzernamen-Routen zu vermeiden. | Navbar und mobile Navbar-Drawer, Nutzerrouten und Servertemplates. | Go-Tests und Template-Rendering ausführen, sobald die Go-Toolchain verfügbar ist. |
