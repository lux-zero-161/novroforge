# Prompt: Gesamte Panel-Funktionalität in WebMain integrieren

Du arbeitest in diesem Repository:

- **Referenzprojekt:** `C:\Users\fates\Documents\GitHub\novroforge\AndereProjekt\panel-release-v2.0.0-beta.2`
- **Zielprojekt:** `C:\Users\fates\Documents\GitHub\novroforge\WebMain`

## Ziel

Übertrage die **gesamte relevante Funktionalität** des Referenzprojekts in `WebMain`, sodass `WebMain` am Ende ein vollständiges Panel mit vergleichbaren Verwaltungsfunktionen wie Pterodactyl bereitstellt. Das Referenzprojekt dient als Funktions- und Verhaltensvorlage. Passe seine Funktionen an die Architektur und Technologien von `WebMain` an, anstatt das Referenzprojekt unverändert hineinzukopieren.

**Das vorhandene Design von WebMain muss unverändert bleiben.** Behalte Layout, Navigation, Branding, Farben, Typografie und die bestehende Benutzerführung bei. Neue Funktionen müssen sich in die vorhandene Oberfläche einfügen. Ändere das Design nicht, um es dem Referenzprojekt oder Pterodactyl anzugleichen.

## Verbindliche Arbeitsregeln

1. Analysiere zunächst beide Projekte gründlich: Technologien, Architektur, vorhandene Funktionen, Datenmodelle, APIs, Authentifizierung, Berechtigungen, UI-Muster, Konfiguration und Tests.
2. Erstelle aus der Analyse eine Funktionsübersicht des Referenzprojekts. Vergleiche jede relevante Funktion mit `WebMain` und arbeite die fehlenden Funktionen vollständig in `WebMain` ein.
3. „Gesamte Funktionalität“ bedeutet: keine relevanten Panel-Funktionen stillschweigend auslassen. Falls eine Funktion technisch nicht sinnvoll oder nicht sicher integrierbar ist, dokumentiere genau, warum, welche Abhängigkeiten fehlen und was für die Umsetzung benötigt wird.
4. Implementiere in `WebMain` mit dessen bestehenden Technologien, Konventionen und Komponenten. Übernimm keine fremden Frameworks oder komplette Projektstrukturen, wenn eine native Integration möglich ist.
5. Das Referenzprojekt ist ausschließlich eine Vorlage und muss unverändert bleiben. Alle Code-, Konfigurations-, Datenbank- und Dokumentationsänderungen gehören in `WebMain`.
6. Bewahre sämtliche bestehenden Funktionen von `WebMain`. Entferne oder überschreibe nichts, was nicht zwingend für die Integration erforderlich ist.
7. Berücksichtige alle zusammengehörigen Bereiche: Benutzeroberfläche, Backend/API, Datenhaltung und Migrationen, Authentifizierung und Autorisierung, Konfiguration, Fehlerbehandlung und Tests. Eine Oberfläche ohne funktionierende Backend-Anbindung gilt nicht als abgeschlossene Funktion.
8. Achte besonders auf Zugriffskontrollen, Eingabevalidierung, sichere Geheimnisbehandlung und verständliche Fehlerausgaben. Verwende vorhandene Sicherheitsmechanismen und führe keine unsicheren Standardberechtigungen ein.
9. Ergänze oder aktualisiere passende Tests und Dokumentation. Führe die verfügbaren Tests, Linter, Formatierer und Builds aus. Behebe Fehler, die durch deine Änderungen verursacht wurden.
10. Arbeite in überschaubaren, überprüfbaren Schritten. Beginne mit der Analyse; nimm danach die Implementierung vor und prüfe sie fortlaufend. Beende die Arbeit nicht nach einem bloßen Plan oder einer Teilimplementierung.

## Erwartete Funktionen

Prüfe das Referenzprojekt und integriere alle dort vorhandenen, relevanten Panel-Funktionen. Dazu können insbesondere gehören:

- Serverübersicht und Serverdetails
- Server erstellen, bearbeiten, verwalten und – soweit vorgesehen – löschen
- Serverstatus, Ressourcenverbrauch und Live-Aktualisierungen
- Konsole und sichere Übermittlung von Befehlen
- Starten, Stoppen und Neustarten
- Serverkonfigurationen, Umgebungen und unterstützte Verwaltungsaktionen
- Benutzer, Rollen, Berechtigungen und Serverzugriffe
- Aktivitätsverlauf, Benachrichtigungen und Fehlerzustände
- Vorhandene Installations-, Datei-, Backup-, Netzwerk- oder Zeitplanfunktionen, sofern sie im Referenzprojekt enthalten und für `WebMain` vorgesehen sind
- Administrations- und Systemeinstellungen, die zum Betrieb des Panels erforderlich sind

Diese Liste ist nicht abschließend. Leite den tatsächlichen Umfang aus dem Quellcode des Referenzprojekts ab und gleiche ihn mit den Möglichkeiten und Grenzen von `WebMain` ab.

## Designvorgabe

Nutze ausschließlich das bestehende Designsystem und die UI-Konventionen von `WebMain`. Ergänze notwendige Ansichten und Bedienelemente, ohne vorhandene Seiten unnötig umzubauen. Wenn eine Funktion keine passende UI-Komponente hat, gestalte sie so zurückhaltend, dass sie sich nahtlos in die bestehende Oberfläche einfügt. Keine Neugestaltung, kein neues Branding und kein Austausch der bestehenden Navigation.

## Abschlussbericht

Fasse nach der Umsetzung knapp und nachvollziehbar zusammen:

- welche Funktionen aus dem Referenzprojekt integriert wurden;
- welche Funktionen nicht integriert werden konnten und warum;
- welche wichtigen Dateien, APIs, Datenbankstrukturen oder Migrationen geändert wurden;
- welche Tests, Linter und Builds ausgeführt wurden und mit welchem Ergebnis;
- welche manuellen Konfigurations- oder Betriebsschritte noch erforderlich sind.

Arbeite direkt im Zielprojekt `WebMain`. Beende die Aufgabe erst, wenn die integrierten Funktionen implementiert und die verfügbaren Prüfungen ausgeführt wurden oder konkrete, nicht selbst lösbare Blocker dokumentiert sind.
