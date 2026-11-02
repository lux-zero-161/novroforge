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

### Gesamtfortschritt: ca. 37 % (Schätzung)

`[####......] 37 %`

- Die Schätzung bezieht sich auf den geplanten Gesamtumfang eines Panel-Integrationsprojekts (Analyse, sichere Wings-Anbindung, Nodes/Server, Benutzerzugriff und weitere Panel-Funktionen), nicht auf die Anzahl der geänderten Dateien. Die vollständige Funktionsinventur und Nutzerberechtigungen fehlen noch.

- [ ] Analyse beider Projekte abgeschlossen
- [ ] Funktionen des Referenzprojekts erfasst und mit `WebMain` abgeglichen
- [ ] Integrationsumfang und Abhängigkeiten dokumentiert
- [ ] Umsetzung abgeschlossen
- [ ] Tests, Linter und Builds ausgeführt
- [ ] Abschlussbericht erstellt

Aktueller Umsetzungsschritt: sichere Wings-Node-Konfiguration, Admin-Verwaltung und Benutzeransicht mit expliziter Serverzuordnung. Node-Tokens werden mit WebMain-Verschlüsselung gespeichert und können nach erfolgreichem Verbindungstest rotiert werden. Admins können Serveranzeigenamen verwalten und getrennte Power-, Konsolen-, Dateibrowser- und Dateivorschau-Rechte vergeben. Zugewiesene Nutzer sehen Status, Ressourcen und Konsolenlogs; diese werden alle fünf Sekunden aktualisiert. Der Dateibrowser ist read-only; kleine UTF-8-Textdateien benötigen eine eigene Vorschau-Freigabe. Dateiänderungen, Downloads und weitere Panel-Funktionen fehlen weiterhin.

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
| --- | --- | --- | --- |
| Wings-Kommunikation: Serverdetails, Logs, Power und Konsole | Ja | In Arbeit | `WebMain/services/wings/client.go`; Status, Suspendierung, CPU, Speicher, Datenträger, Laufzeit, Netzwerk und letzte 100 Logs in der Nutzeransicht alle 5 Sekunden aktualisiert; Power und Befehle sind verfügbar |
| Nodes in der Admin-Oberfläche registrieren, Verbindung testen, Token rotieren und entfernen | Ja | In Arbeit | `WebMain/models/wings/node.go`, Migration `v18b_add-wings-nodes.go`, `/admin/wings/nodes`; Token wird verschlüsselt gespeichert und vor Rotation mit Wings geprüft |
| Serverübersicht und Serverdetails | Ja | In Arbeit | Adminübersicht und Nutzeransicht unter `/user/servers`, nur explizit zugewiesene Server; Status und Ressourcen aktualisieren sich alle 5 Sekunden |
| Server erstellen, bearbeiten und verwalten | Ja | In Arbeit | Registrierung vorhandener Wings-Server; Admins können den WebMain-Anzeigenamen ändern. Provisionierung, technische Konfiguration und weitere Bearbeitungsfunktionen fehlen. |
| Status und Ressourcenübersicht | Ja | In Arbeit | Nutzeransicht zeigt Zustand, Suspendierung, CPU, Speicher, Datenträger, Laufzeit sowie empfangenen/gesendeten Netzwerkverkehr; Werte werden alle 5 Sekunden neu geladen |
| Live-Konsole und Befehle | Ja | In Arbeit | Letzte 100 Konsolenzeilen werden alle 5 Sekunden aktualisiert; echtes WebSocket-/Live-Streaming fehlt. Befehle nur für explizit berechtigte Nutzer oder Administratoren |
| Starten, Stoppen und Neustarten | Ja | In Arbeit | Power-Aktionen nur für explizit berechtigte Nutzer oder Administratoren |
| Konfigurationen und Umgebungen | Ja | Noch zu prüfen | |
| Benutzer, Rollen und Berechtigungen | Ja | In Arbeit | Admin kann WebMain-Nutzern Serverzugriff sowie getrennte Power-, Konsolen-, Dateibrowser- und Dateivorschau-Rechte vergeben oder entziehen; vollständiges Rollenmodell fehlt |
| Aktivitätsverlauf und Benachrichtigungen | Ja | Noch zu prüfen | |
| Dateien, Backups, Datenbanken, Netzwerke und Zeitpläne | Ja | In Arbeit | Read-only Dateibrowser für Verzeichnislisten/Metadaten und separate Textvorschau bis 1 MiB unter `/user/servers/{id}/files`; Vorschau benötigt eine zusätzliche Admin-Freigabe. Download, Upload, Bearbeitung und übrige Bereiche fehlen. |
| Nodes, Allocations und Eggs | Ja | Noch zu prüfen | |
| Administration und Systemeinstellungen | Ja | Noch zu prüfen | |
| Weitere relevante Funktionen | Noch zu prüfen | Noch zu prüfen | |

Mögliche Statuswerte: **Noch zu prüfen**, **Offen**, **In Arbeit**, **Erledigt**, **Nicht umgesetzt**.

## Prüfungen

| Prüfung | Ergebnis | Hinweise |
| --- | --- | --- |
| Tests | Teilweise nicht ausführbar | Go-Tests einschließlich Wings-Verzeichnis-, Datei-Pfad- und Vorschau-Limits konnten wegen der fehlenden Go-Toolchain nicht ausgeführt werden. |
| Linter / Formatierer | Teilweise bestanden | ESLint für Wings-Konsole und JS-Einstieg sowie Markdown-Lint bestanden; `gofmt` ist mangels Go-Toolchain nicht ausführbar. |
| Localization JSON | Bestanden | `options/locale_next/locale_en-US.json` mit PowerShell-JSON-Parser eingelesen. |
| Git whitespace check | Bestanden | `git diff --check` meldet keine Fehler. |
| Build / Typprüfung | Teilweise bestanden | Isolierter Webpack-Frontend-Build nach Ergänzung der Ressourcenanzeige und des Dateibrowsers bestanden; Go-Build und Go-Typprüfung sind mangels Go-Toolchain offen. |

## Blocker und nicht umgesetzte Funktionen

- **Noch offen:** Dateiänderungen, Datei-Downloads sowie Backups, Datenbanken und weitere Panel-Bereiche. Dateilisten und begrenzte Textvorschau sind read-only und erfordern getrennte Freigaben.
- Die aktuelle Umgebung stellt `go` und `gofmt` nicht bereit. Daher konnten die hinzugefügten Go-Änderungen hier noch nicht kompiliert oder getestet werden.
- Benutzerkonten können Servern explizit zugeordnet werden. Status/Logs sind für zugewiesene Nutzer lesbar; Power-Aktionen und Konsolenbefehle sind per separater Freigabe geschützt.

## Fortschrittsprotokoll

| Datum | Änderung / Ergebnis | Geänderte Bereiche | Nächster Schritt |
| --- | --- | --- | --- |
| 2026-11-02 | Erste Architekturprüfung: Go/Vue-Zielprojekt und Laravel/React-Referenzpanel festgestellt; Integrationsumfang ist groß und muss phasenweise umgesetzt werden. | Beide Projekte read-only geprüft; Fortschrittsdokument aktualisiert. | WebMain-Funktionen und Erweiterungspunkte vollständig inventarisieren, danach Integrationsschritte festlegen. |
| 2026-11-02 | Wings-Routen und Authentifizierung geprüft: Serveraktionen benötigen das Bearer-Token der jeweiligen Node; WebMain hat in der ersten Suche keine Panel-Integration gefunden. | Wings-Router und Pterodactyl-Daemon-Client read-only geprüft; Fortschrittsdokument aktualisiert. | Sichere Node-/Server-Datenhaltung und Berechtigungsprüfung entwerfen, dann einen getesteten Kernbereich in WebMain implementieren. |
| 2026-11-02 | Wings-HTTP-Client ergänzt: HTTPS für externe Nodes, HTTP nur für Loopback, keine Redirects, begrenzte Antworten und validierte UUIDs/Aktionen. Go-Tests und Formatierung konnten wegen fehlender Go-Toolchain nicht ausgeführt werden. | `WebMain/services/wings/client.go`, `WebMain/services/wings/client_test.go` | Go-Toolchain bereitstellen, Tests/Formatierung ausführen, anschließend verschlüsselte Node-Verwaltung und autorisierte WebMain-Endpunkte ergänzen. |
| 2026-11-02 | Admin-Verwaltung für Wings-Nodes ergänzt: Migration und Modell mit XChaCha20-Poly1305-Verschlüsselung, HTTPS-Validierung sowie Registrierung/Löschung unter `/admin/wings/nodes`. Token wird nicht in der Liste angezeigt. | Modelle, Migration, Admin-Handler/Routen, Admin-Navigation und `wings_nodes.tmpl`; Locale-JSON validiert und `git diff --check` bestanden. | Go-Toolchain bereitstellen und alle Änderungen kompilieren/testen; danach Wings-Serverzuordnung und Rechteprüfung aufbauen. |
| 2026-11-02 | Admin-Verbindungstest ergänzt: authentifizierter Aufruf von Wings `/api/system`; Fehler werden nutzerverständlich angezeigt, Token wird nicht ausgegeben. Requests und Konsolenbefehle im Client zusätzlich begrenzt. | Wings-Client, Node-Modell, Admin-Handler/Routen und Nodes-Ansicht. | Erneute Build-/Testvalidierung nach Verfügbarkeit von Go; anschließend sichere Serverzuordnung und Rechteprüfung. |
| 2026-11-02 | Benutzerzuordnung ergänzt: Admins können aktive WebMain-Konten einem Server zuweisen und Power-/Konsolenrechte separat vergeben. Nutzer sehen unter `/user/servers` nur ihre zugewiesenen Server und können je nach Freigabe Aktionen ausführen. Beim Entfernen einer Serverregistrierung werden ihre Zugriffszuordnungen in einer Transaktion gelöscht. Fortschritt auf ca. 30 % aktualisiert. | Wings-Zugriffsmodell und Migration `v18d`, Admin-/Nutzerrouten, Berechtigungsprüfungen, Benutzeransichten, Navigation und Locale. | Go-Toolchain für Build/Tests bereitstellen; danach Live-Konsole, Datei- und Backup-Funktionen sowie weitere Panel-Bereiche implementieren. |
| 2026-11-02 | Nutzer-Serverseiten in die normale Desktop- und mobile Navigation aufgenommen; Serverrouten liegen unter `/user/servers`, um Konflikte mit öffentlichen Benutzernamen-Routen zu vermeiden. | Navbar und mobile Navbar-Drawer, Nutzerrouten und Servertemplates. | Go-Tests und Template-Rendering ausführen, sobald die Go-Toolchain verfügbar ist. |
| 2026-11-02 | Nutzerkonsole ergänzt: Die Seite fragt die letzten 100 Logs im Fünf-Sekunden-Takt über eine authentifizierte, zugriffsgeprüfte Route ab. Seiteninhalte werden als Text gesetzt, Antworten nicht im Browser-Cache gespeichert und Fehler sichtbar mit erneutem Versuch behandelt. Der gezielte ESLint-Lauf und die Locale-Prüfung sind erfolgreich. | `routers/web/user/wings.go`, Nutzerrouten, Servertemplate, `web_src/js/features/wings-console.js`, JS-Einstieg und Locale. | Frontend-Build/Go-Tests ausführen, sobald Toolchains bereitstehen; anschließend WS-Livekonsole und weitere Panel-Funktionen. |
| 2026-11-02 | Admins können das gespeicherte Wings-Node-Token ersetzen. Die Verbindung mit dem neuen Token wird vor dem Speichern geprüft; anschließend wird es mit dem bestehenden Knotenschlüssel verschlüsselt. Das Token wird nicht wieder angezeigt. | Node-Modell, Admin-Route und Formular, englische Locale. | Go-Tests und Formatierung ausführen, sobald die Go-Toolchain verfügbar ist; danach Datei-/Backup-Funktionen priorisieren. |
| 2026-11-02 | Die geschützte Nutzeraktualisierung liefert neben den letzten 100 Logs auch Wings-Status und Ressourcenwerte. Status, CPU, Speicher und Datenträger werden mit der Konsole alle fünf Sekunden aktualisiert; Fehlerstatus und Antwortschema sind geprüft. | Nutzerhandler und JSON-Antwort, Servertemplate, Konsolen-JavaScript, englische Locale. | Frontend-Lint/Build prüfen; Go-Tests bleiben bis zur Verfügbarkeit der Go-Toolchain offen. |
| 2026-11-02 | Admins können den Anzeigenamen einer registrierten Wings-Serverinstanz in WebMain umbenennen. Die Eingabe wird normalisiert und auf 1 bis 64 Zeichen begrenzt; Wings selbst wird nicht geändert. Grenzfalltests für die Namensvalidierung ergänzt. | Wings-Servermodell und Test, Admin-Handler/Route/Formular sowie englische Locale. | Go-Tests ausführen, sobald die Go-Toolchain verfügbar ist; danach Datei-/Backup-Funktionen planen. |
| 2026-11-02 | Serverdetails erweitert: Nutzer sehen zusätzlich Suspendierungsstatus, Laufzeit und empfangene/gesendete Netzwerkbytes. Die Werte werden mit Status, CPU, Speicher, Datenträger und Konsole im Fünf-Sekunden-Takt aktualisiert. | Nutzer-Servertemplate, Live-Aktualisierung und englische Locale. | Frontend-Lint/Build erneut prüfen; Go-Tests bleiben bis zur Verfügbarkeit der Go-Toolchain offen. |
| 2026-11-02 | Read-only Dateibrowser ergänzt: Wings-Verzeichniseinträge und Metadaten werden angezeigt; Pfade werden bereinigt und Traversal-Versuche abgewiesen. Der Zugriff benötigt ein neues, standardmäßig deaktiviertes Serverrecht; Migration ergänzt dieses Recht für bestehende Zuordnungen. Symlink-Verzeichnisse sind nicht navigierbar. | Wings-Client und Tests, Migration `v18e`, Zugriffsmodell/Admin-Freigabe, Nutzerroute und Dateiliste-Template, Locale. | Go-Tests/Migration prüfen, sobald die Go-Toolchain verfügbar ist; danach Datei-Downloads und Änderungen nur mit expliziten Rechten planen. |
| 2026-11-02 | Read-only Textvorschau ergänzt: separates, standardmäßig verweigertes Recht; nur UTF-8-Text mit unterstütztem MIME-Typ und maximal 1 MiB. Pfade werden geprüft, Symlinks nicht als Vorschau verlinkt, Inhalte im HTML escaped und Antworten nicht gecacht. | Wings-Client und Tests, Migration `v18f`, Zugriffsmodell/Admin-Freigabe, Nutzerroute/Templates, Locale. Gezielter ESLint, Markdownlint, Locale-Prüfung, isolierter Produktions-Webpack-Build und `git diff --check` erfolgreich. | Go-Tests sowie Template-/Migrationsvalidierung nach Bereitstellung der Go-Toolchain ausführen; anschließend Downloads/Änderungen getrennt planen. |
