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

- **Gesamtfortschritt: ca. 20 % (Schätzung)** — `[##........]`
- Die Schätzung bezieht sich auf den geplanten Gesamtumfang eines Panel-Integrationsprojekts (Analyse, sichere Wings-Anbindung, Nodes/Server, Benutzerzugriff und weitere Panel-Funktionen), nicht auf die Anzahl der geänderten Dateien. Die vollständige Funktionsinventur und Nutzerberechtigungen fehlen noch.

- [ ] Analyse beider Projekte abgeschlossen
- [ ] Funktionen des Referenzprojekts erfasst und mit `WebMain` abgeglichen
- [ ] Integrationsumfang und Abhängigkeiten dokumentiert
- [ ] Umsetzung abgeschlossen
- [ ] Tests, Linter und Builds ausgeführt
- [ ] Abschlussbericht erstellt

Aktueller Umsetzungsschritt: sichere Wings-Node-Konfiguration sowie eine Admin-Verwaltung für vorhandene Server. Node-Tokens werden mit WebMain-Verschlüsselung gespeichert. Server können einer Node zugeordnet, auf Status geprüft und durch Administratoren gestartet, gestoppt, neugestartet oder mit Konsolenbefehlen angesprochen werden. Benutzerzuordnungen, Zugriffsrechte, Live-Konsole und Nutzeroberfläche für Server fehlen; das ist noch kein vollständiges Serverpanel.

## Erste Architekturprüfung

- `WebMain` ist ein Go-Projekt (Go 1.26) mit einer bestehenden Vue-3-Weboberfläche und eigener Router-, Modell- und API-Struktur.
- Das Referenzprojekt ist Pterodactyl Panel 2.0: Backend mit PHP 8.3 und Laravel 13, Frontend mit React 19; es besitzt umfangreiche eigene APIs und Datenmodelle.
- Im Referenzprojekt sind unter anderem Client-, Application- und Admin-APIs sowie Serveraktionen für Konsole, Leistung, Dateien, Backups, Datenbanken, Zeitpläne, Nodes, Eggs, Benutzer und Berechtigungen vorhanden.
- Eine passende Pterodactyl-/Wings-Integration wurde bei der ersten Suche in `WebMain` nicht gefunden. Die vollständige Prüfung des Zielprojekts steht noch aus.
- Wegen der unterschiedlichen Architekturen ist eine vollständige Integration eine umfangreiche, schrittweise Neuumsetzung. Das Referenzprojekt darf nicht einfach in `WebMain` kopiert werden; das bestehende WebMain-Design bleibt unverändert.
- Das vorhandene `Wings`-Projekt stellt Server-Endpunkte für Status, Konsole, Power-Aktionen und weitere Ressourcen bereit. Administrative Wings-Aufrufe erfordern ein konfiguriertes Bearer-Token pro Node.
- Vor Beginn dieser Integration hatte WebMain weder eine Wings-Node-Konfiguration noch Server-Zugriffsmodelle für Panel-Benutzer. Die Node-Konfiguration ist nun implementiert; Zuordnungen zu Servern und Benutzerberechtigungen fehlen weiterhin. Es werden keine echten Tokens in Quellcode oder Dokumentation eingetragen.
- Node- und Serververwaltung sind derzeit ausschließlich im Admin-Bereich verfügbar; sie geben normalen Benutzern noch keinen Zugriff auf Server.

## Funktions- und Umsetzungsübersicht

| Bereich/Funktion | Im Referenzprojekt vorhanden | Status in WebMain | Umsetzung / Datei / Hinweis |
|---|---|---|---|
| Wings-Kommunikation: Serverdetails, Logs, Power und Konsole | Ja | In Arbeit | `WebMain/services/wings/client.go`; Serverstatus, Power und Befehle in Admin-Handler eingebunden; Logs noch nicht in einer Live-Konsole dargestellt |
| Nodes in der Admin-Oberfläche registrieren, Verbindung testen und entfernen | Ja | In Arbeit | `WebMain/models/wings/node.go`, Migration `v18b_add-wings-nodes.go`, `/admin/wings/nodes`; Token wird verschlüsselt gespeichert |
| Serverübersicht und Serverdetails | Ja | In Arbeit (Admin) | Registrierung vorhandener Wings-Server, Statusabfrage und Übersicht; keine Nutzeransicht |
| Server erstellen, bearbeiten und verwalten | Ja | Noch zu prüfen | |
| Status und Ressourcenübersicht | Ja | Noch zu prüfen | |
| Live-Konsole und Befehle | Ja | In Arbeit (Admin) | Konsolenbefehle können gesendet werden; Live-Log/Streaming-Konsole fehlt |
| Starten, Stoppen und Neustarten | Ja | In Arbeit (Admin) | Admin-Aktionen an Wings; keine Benutzerrechte/Serverzuordnung |
| Konfigurationen und Umgebungen | Ja | Noch zu prüfen | |
| Benutzer, Rollen und Berechtigungen | Ja | Noch zu prüfen | |
| Aktivitätsverlauf und Benachrichtigungen | Ja | Noch zu prüfen | |
| Dateien, Backups, Datenbanken, Netzwerke und Zeitpläne | Ja | Noch zu prüfen | |
| Nodes, Allocations und Eggs | Ja | Noch zu prüfen | |
| Administration und Systemeinstellungen | Ja | Noch zu prüfen | |
| Weitere relevante Funktionen | Noch zu prüfen | Noch zu prüfen | |

Mögliche Statuswerte: **Noch zu prüfen**, **Offen**, **In Arbeit**, **Erledigt**, **Nicht umgesetzt**.

## Prüfungen

| Prüfung | Ergebnis | Hinweise |
|---|---|---|
| Tests | Nicht ausführbar | `go test ./services/wings` wurde versucht; Go-Toolchain ist in der aktuellen Shell nicht installiert/verfügbar. |
| Linter / Formatierer | Nicht ausführbar | `gofmt` wurde versucht; Go-Toolchain ist in der aktuellen Shell nicht installiert/verfügbar. |
| Localization JSON | Bestanden | `options/locale_next/locale_en-US.json` mit PowerShell-JSON-Parser eingelesen. |
| Git whitespace check | Bestanden | `git diff --check` meldet keine Fehler. |
| Build / Typprüfung | Noch nicht ausgeführt | Nach Bereitstellung der Go-Toolchain erforderlich. |

## Blocker und nicht umgesetzte Funktionen

- **Noch offen:** Rotation der Wings-Node-Tokens sowie Zuordnung von Nodes/Servern zu WebMain-Benutzern und deren Berechtigungen.
- Die aktuelle Umgebung stellt `go` und `gofmt` nicht bereit. Daher konnten die hinzugefügten Go-Änderungen hier noch nicht kompiliert oder getestet werden.
- Die aktuelle Node-Verwaltung speichert nun Wings-Servermetadaten. Benutzerkonten sind jedoch noch nicht einzelnen Servern zugeordnet; Serveraktionen sind deshalb vorerst nur für Administratoren verfügbar.

## Fortschrittsprotokoll

| Datum | Änderung / Ergebnis | Geänderte Bereiche | Nächster Schritt |
|---|---|---|---|
| 2026-11-02 | Erste Architekturprüfung: Go/Vue-Zielprojekt und Laravel/React-Referenzpanel festgestellt; Integrationsumfang ist groß und muss phasenweise umgesetzt werden. | Beide Projekte read-only geprüft; Fortschrittsdokument aktualisiert. | WebMain-Funktionen und Erweiterungspunkte vollständig inventarisieren, danach Integrationsschritte festlegen. |
| 2026-11-02 | Wings-Routen und Authentifizierung geprüft: Serveraktionen benötigen das Bearer-Token der jeweiligen Node; WebMain hat in der ersten Suche keine Panel-Integration gefunden. | Wings-Router und Pterodactyl-Daemon-Client read-only geprüft; Fortschrittsdokument aktualisiert. | Sichere Node-/Server-Datenhaltung und Berechtigungsprüfung entwerfen, dann einen getesteten Kernbereich in WebMain implementieren. |
| 2026-11-02 | Wings-HTTP-Client ergänzt: HTTPS für externe Nodes, HTTP nur für Loopback, keine Redirects, begrenzte Antworten und validierte UUIDs/Aktionen. Go-Tests und Formatierung konnten wegen fehlender Go-Toolchain nicht ausgeführt werden. | `WebMain/services/wings/client.go`, `WebMain/services/wings/client_test.go` | Go-Toolchain bereitstellen, Tests/Formatierung ausführen, anschließend verschlüsselte Node-Verwaltung und autorisierte WebMain-Endpunkte ergänzen. |
| 2026-11-02 | Admin-Verwaltung für Wings-Nodes ergänzt: Migration und Modell mit XChaCha20-Poly1305-Verschlüsselung, HTTPS-Validierung sowie Registrierung/Löschung unter `/admin/wings/nodes`. Token wird nicht in der Liste angezeigt. | Modelle, Migration, Admin-Handler/Routen, Admin-Navigation und `wings_nodes.tmpl`; Locale-JSON validiert und `git diff --check` bestanden. | Go-Toolchain bereitstellen und alle Änderungen kompilieren/testen; danach Wings-Serverzuordnung und Rechteprüfung aufbauen. |
| 2026-11-02 | Admin-Verbindungstest ergänzt: authentifizierter Aufruf von Wings `/api/system`; Fehler werden nutzerverständlich angezeigt, Token wird nicht ausgegeben. Requests und Konsolenbefehle im Client zusätzlich begrenzt. | Wings-Client, Node-Modell, Admin-Handler/Routen und Nodes-Ansicht. | Erneute Build-/Testvalidierung nach Verfügbarkeit von Go; anschließend sichere Serverzuordnung und Rechteprüfung. |
| Weiterführung | Serverseitige Zuordnung vorhandener Wings-Server zu Nodes ergänzt; Registrierung prüft den Server über Wings. Admin-Oberfläche bietet Statusprüfung, Start/Stop/Restart und Befehle. Node-Konfiguration kann bei zugeordneten Servern nicht versehentlich gelöscht werden. Gesamtfortschritt auf ca. 20 % geschätzt. | Wings-Servermodell und Migration, Admin-Handler/Routen, Nodes-/Serveransicht und englische Locale. | Go-Toolchain bereitstellen und kompilieren/testen; danach Benutzerzuordnung, Rechteprüfung und eine Nutzer-Serveransicht entwickeln. |
