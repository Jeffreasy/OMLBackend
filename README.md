OdomosML
OdomosML is een RESTful API geschreven in Go (Golang) met Gin als webframework. Het project demonstreert:

- Authenticatie en autorisatie via JWT-tokens (JSON Web Tokens)
- Gebruikersbeheer (aanmaken, bewerken, verwijderen)
- Klantenbeheer (CRUD-operaties)
- Audit logging (registratie van acties, wijzigingen, en verwijderingen in de database)

Inhoudsopgave
- Project Overzicht
- Belangrijkste Features
- Structuur en Mappen
- Installatie & Setup
- Environment Variables
- Uitvoeren
- API Endpoints
- Authenticatie & Autorisatie
- Audit Logging
- Productie Overwegingen
- TODO's / Verbeterpunten
- Bijdragen
- Licentie

Project Overzicht
Dit project is bedoeld als voorbeeld van een schaalbare en onderhoudbare Go-applicatie die gebruikmaakt van:

- Gin: lichtgewicht HTTP-server en router
- GORM: ORM voor database-interacties
- PostgreSQL: relationele database
- JWT: voor login en session management

We hanteren een “clean architecture”-achtige indeling, met mappen voor delivery (handlers), service, repository en model (data-objecten en logica).

Belangrijkste Features

Gebruikersbeheer:
- Aanmaken en updaten van gebruikers (inclusief wachtwoord-hashing).
- Autorisatie-rollen: admin, user, etc.

Klantenbeheer:
- CRUD-operaties (Create, Read, Update, Delete).
- Mogelijkheid tot gefilterde zoekopdrachten (paginering, zoekterm, sortering).

Authenticatie:
- JWT-token genereren bij inloggen of registreren.
- Token refresh endpoint.

Auditing:
- Alle mutaties (POST, PUT, PATCH, DELETE) worden vastgelegd in audit logs.
- Mogelijkheid om audit logs gefilterd op te vragen (per entiteit, actie, datum, enz.).

Structuur en Mappen

De code is georganiseerd in diverse mappen en bestanden:

```bash
.
├── config/
│   └── config.go            # Laden van omgevingsvariabelen (DB, server-poort, JWT-secret, etc.)
├── internal/
│   ├── audit/
│   │   ├── delivery/http/   # AuditHandler (Gin handlert)
│   │   ├── model/           # AuditLog en gerelateerde structs
│   │   ├── repository/      # AuditRepository (DB-operaties)
│   │   └── service/         # AuditService (bedrijfslogica voor logging)
│   ├── auth/
│   │   ├── delivery/http/   # AuthHandler (login, register, refresh)
│   │   ├── model/           # Auth gerelateerde models (TokenResponse, Claims, etc.)
│   │   └── service/         # AuthService (JWT genereren/valideren, registratie)
│   ├── customer/
│   │   ├── delivery/http/   # CustomerHandler (klantenbeheer)
│   │   ├── model/           # Customer data (name, email, etc.)
│   │   ├── repository/      # CustomerRepository (DB-operaties)
│   │   └── service/         # CustomerService (businesslogica)
│   ├── user/
│   │   ├── delivery/http/   # UserHandler (gebruikersbeheer)
│   │   ├── model/           # User model + rollen
│   │   ├── repository/      # UserRepository (DB-operaties)
│   │   └── service/         # UserService (businesslogica)
│   └── middleware/          # AuthMiddleware, RoleMiddleware, AuditMiddleware
├── pkg/
│   └── database/
│       └── database.go      # Database connectie (PostgreSQL via GORM)
├── go.mod                    # Go modules + dependencies
├── go.sum                    # Dependency checksums
└── main.go                   # Startpunt van de applicatie (laadt config, router, etc.)
```

Installatie & Setup

Vereisten
- Go >= 1.18
- PostgreSQL (optioneel Docker als je snel een test-DB wilt draaien)
- GIT voor het klonen van de repository

Stappen

Repository klonen:

```bash
git clone <URL-naar-jouw-repo> odomosml
cd odomosml
```

Go modules installeren:

```bash
go mod tidy
```
Dit haalt alle benodigde pakketten binnen.

Database configuratie:

- Creëer een PostgreSQL-database (bijv. odomosml).
- Controleer de variabelen in Environment Variables of pas .env aan.
- Eventueel .env-bestand maken:
  - Als je met een .env-bestand werkt, maak deze in de root aan.
  - Vul daar je DB-gegevens in, server-poort, etc.

Environment Variables

Het project haalt diverse instellingen op uit omgevingsvariabelen. Dit gebeurt in **config/config.go**. De belangrijkste:

| Variabele       | Default         | Beschrijving                                              |
|-----------------|-----------------|-----------------------------------------------------------|
| SERVER_ADDRESS  | :8080           | Poort waarop de server luistert                           |
| DB_HOST         | localhost       | Hostname van de DB                                        |
| DB_PORT         | 5432            | Port van de DB                                            |
| DB_USER         | postgres        | Database gebruikersnaam                                   |
| DB_PASSWORD     | Bootje@12       | Database wachtwoord                                       |
| DB_NAME         | odomosml        | Database naam                                             |
| JWT_SECRET      | your-secret-key | Secret key voor JWT-signing. In productie: zet dit op iets sterks! |

*Let op:* de standaard DBPassword in code is alleen voor test/demo. In productie gebruik je een sterke, geheime variabele.

Uitvoeren

Je kunt de applicatie starten met:

```bash
go run main.go
```
Bij het opstarten worden alle tabellen in je PostgreSQL-database gedropt en opnieuw aangemaakt. Er wordt een admin-gebruiker aangemaakt:

- username: admin
- email: admin@example.com
- password: admin123

Na succesvolle start luistert de applicatie op het adres dat staat in **SERVER_ADDRESS** (standaard :8080).

Open een REST client (bijv. Postman of Insomnia) en maak requests naar http://localhost:8080/.

API Endpoints

Hier een overzicht van de belangrijkste routes. Prefix: **/api**.

Auth
Methode	Endpoint	        Beschrijving	                                  Bescherming
POST	    /auth/login	    Inloggen met email + wachtwoord	              Publiek
POST	    /auth/register	Nieuw account aanmaken	                      Publiek
POST	    /auth/refresh	    Nieuw JWT token op basis van huidig token	      JWT vereist, alle rollen

Voorbeeld (login):

```json
POST /api/auth/login
{
  "email": "admin@example.com",
  "password": "admin123"
}
```

Response (succes):

```json
{
  "access_token": "<JWT-HIER>",
  "token_type": "Bearer",
  "expires_in": 86400,
  "username": "admin",
  "email": "admin@example.com",
  "role": "ADMIN"
}
```

Users
Methode	Endpoint	        Beschrijving	                    Vereiste rol
GET	    /users	        Haal alle gebruikers op	            ADMIN
GET	    /users/:id	    Haal gebruiker op ID	            ADMIN
POST	    /users	        Maak nieuwe gebruiker aan	        ADMIN
PUT	    /users/:id	    Vervang bestaande gebruiker	    ADMIN
DELETE	    /users/:id	    Verwijder bestaande gebruiker	    ADMIN

Klanten
Methode	Endpoint	        Beschrijving	                                      Vereiste rol
GET	    /klanten	    Lijst van klanten (met zoek, paging, etc.)	      ADMIN of USER
GET	    /klanten/:id	Specifieke klant op ID	                              ADMIN of USER
POST	    /klanten	    Nieuwe klant aanmaken	                              ADMIN of USER
PUT	    /klanten/:id	Bestaande klant volledig updaten	                  ADMIN of USER
PATCH	    /klanten/:id	Bestaande klant gedeeltelijk updaten	              ADMIN of USER
DELETE	    /klanten/:id	Klant verwijderen	                                  ADMIN of USER

Voorbeeld (klant maken):

```json
POST /api/klanten
{
  "name": "Test Klant",
  "email": "test@klant.nl",
  "phone": "+31 6 12345678",
  "address": "Straat 123, 5678AB Plaats"
}
```

Audit Logs
Methode	Endpoint	        Beschrijving	                                        Vereiste rol
GET	    /logs	        Audit-logs ophalen (met diverse filters / paginatie)	    ADMIN

Beschikbare query-params:
page, pageSize, entityType=klanten|users, actionType=create|update|delete, startDate=<ISO8601>, endDate=<ISO8601>

Authenticatie & Autorisatie

Authenticatie vindt plaats via JWT-tokens. Je krijgt een token bij **POST /auth/login** of **POST /auth/register**.  
Autorisatie wordt afgedwongen via de AuthMiddleware en de RoleMiddleware in Gin.  
Om toegang te krijgen tot beschermde endpoints, stuur je in je header:

```makefile
Authorization: Bearer <JWT>
```

*Opgelet:* De standaard Refresh-functionaliteit maakt nog gebruik van een shortcut (Login(email, "")). Overweeg een aparte functie in de service voor echte tokenrefresh (zie TODO).

Audit Logging

AuditMiddleware verwerkt elke POST, PUT, PATCH en DELETE request.  
De gewijzigde/nieuwe data wordt met de oude data vergeleken, en het verschil wordt opgeslagen in de **audit_logs**-tabel.  
Via **GET /api/logs** (alleen voor admins) kun je de audit logs inzien.  
**LET OP:** GET-requests worden momenteel niet gelogd. Als je “read”-acties ook wilt tracken, kun je dat aanpassen in de middleware door ActionRead te activeren.

Productie Overwegingen

- Drop and Create Tables: Standaard worden bij elke start alle tabellen gedropt. Maak dit optioneel met een environment-variabele (zoals DB_DROP_TABLES) zodat in productie je data niet gewist wordt.
- Wachtwoord-hashing bij updates: De code heeft een BeforeCreate-hook, maar nog niet BeforeUpdate. Implementeer BeforeSave om ook bij updates het wachtwoord automatisch te hashen.
- JWT-secret: Gebruik een veilige, lang random secret in productie, nooit de default "your-secret-key".
- SSLMode: In productie wil je PostgreSQL gewoonlijk met SSL aanroepen (sslmode=require of vergelijkbaar).
- Logging: Overweeg structured logging (bijv. logrus of zap) voor meer controle en consistentie.

TODO's / Verbeterpunten

**Token Refresh Flow**
- Momenteel wordt Login(userClaims.Email, "") aangeroepen. Splits dit uit in een RefreshToken() methode op de AuthService.

**Wachtwoord-hashing bij Update**
- Implementeer een BeforeSave-hook of BeforeUpdate-hook in het User model, zodat nieuwe wachtwoorden automatisch gehasht worden.

**Optionele Table-Drop**
- Maak het droppen van tabellen optioneel; in productie wil je data behouden.
- Voeg bijvoorbeeld DB_DROP_TABLES=true toe als env en check deze in database.NewPostgresDB().

**Expand Logging**
- Voeg READ logging toe in AuditMiddleware als je ook GET-verzoeken wilt vastleggen.

**Deployment Scripts / Docker**
- Schrijf Dockerfiles of Kubernetes-manifests om het project in containers uit te rollen.
- Optimaliseer build stages (multi-stage builds) voor productie.

**Unit Tests / Integration Tests**
- Voeg tests toe voor je handlers, services, en repos.
- Maak gebruik van mocking of testcontainers (bijv. testcontainers-go).

Bijdragen

Bijdragen in de vorm van pull requests en feature requests zijn welkom!  
Volg hiervoor de standaard GitHub flow:
- Fork deze repository
- Maak een featurebranch: `git checkout -b feature/naam-van-feature`
- Codeer en commit je wijzigingen
- Dien een pull request in op de main branch

Licentie

Dit project is beschikbaar onder de MIT-licentie. Je mag de code vrij gebruiken, aanpassen en distribueren zolang je de licentie respecteert.
