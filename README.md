# OdomosML Backend

Backend API voor het OdomosML platform.

## Inhoudsopgave

- [Installatie](#installatie)
- [Configuratie](#configuratie)
- [Ontwikkeling](#ontwikkeling)
- [API Endpoints](#api-endpoints)
- [Best Practices](#best-practices)
- [Code Review Opmerkingen](#code-review-opmerkingen)

## Installatie

### Vereisten

- Go 1.18 of hoger
- PostgreSQL 13 of hoger

### Stappen

1. Clone de repository:
   ```bash
   git clone https://github.com/yourusername/odomosml.git
   cd odomosml
   ```

2. Installeer dependencies:
   ```bash
   go mod download
   ```

3. Maak een `.env` bestand aan (zie `.env.example` voor een voorbeeld):
   ```bash
   cp .env.example .env
   ```

4. Start de applicatie:
   ```bash
   go run cmd/omlbackend/main.go
   ```

## Configuratie

De applicatie gebruikt environment variabelen voor configuratie. Zie `.env.example` voor alle beschikbare opties.

Belangrijke configuratie opties:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: Database configuratie
- `JWT_SECRET`: Secret key voor JWT tokens (verander dit in productie!)
- `SERVER_ADDRESS`: Adres waarop de server draait (default: `:8080`)
- `DB_DROP_TABLES`: Zet op `true` om tabellen te droppen bij startup (alleen voor development!)

## Ontwikkeling

### Project Structuur

```
OMLBackend/
├── cmd/
│   └── omlbackend/
│       └── main.go           # Entry point
├── config/
│   └── config.go             # Configuratie
├── internal/
│   ├── app/
│   │   └── app.go            # App setup
│   ├── audit/                # Audit logging
│   ├── auth/                 # Authenticatie
│   ├── customer/             # Klantenbeheer
│   ├── middleware/           # Middleware
│   └── user/                 # Gebruikersbeheer
├── pkg/
│   └── database/             # Database helpers
├── .env.example              # Voorbeeld configuratie
├── go.mod                    # Go modules
└── README.md                 # Deze file
```

### Architectuur

De applicatie volgt een Clean Architecture patroon:

- **Models**: Datastructuren en business regels
- **Repositories**: Data access layer
- **Services**: Business logic
- **Handlers**: HTTP request handlers

## API Endpoints

### Authenticatie

- `POST /api/auth/login`: Inloggen
- `POST /api/auth/register`: Registreren
- `POST /api/auth/refresh`: Token vernieuwen

### Gebruikers

- `GET /api/users`: Alle gebruikers ophalen
- `GET /api/users/:id`: Gebruiker ophalen
- `POST /api/users`: Gebruiker aanmaken
- `PUT /api/users/:id`: Gebruiker bijwerken
- `DELETE /api/users/:id`: Gebruiker verwijderen

### Klanten

- `GET /api/klanten`: Alle klanten ophalen
- `GET /api/klanten/:id`: Klant ophalen
- `POST /api/klanten`: Klant aanmaken
- `PUT /api/klanten/:id`: Klant bijwerken
- `PATCH /api/klanten/:id`: Klant gedeeltelijk bijwerken
- `DELETE /api/klanten/:id`: Klant verwijderen

### Audit Logs

- `GET /api/logs`: Audit logs ophalen

## Best Practices

### Beveiliging

- JWT tokens hebben een expiratie tijd van 24 uur
- Wachtwoorden worden gehasht met bcrypt
- Rollen worden gecontroleerd voor elke beschermde route
- Audit logging voor alle mutaties

### Performance

- Database indexen voor veelgebruikte velden
- Trigram indexen voor ILIKE zoekopdrachten
- Paginering voor alle lijst endpoints
- Maximale pageSize om database overbelasting te voorkomen

### Error Handling

- Consistente error responses met `success` veld
- Duidelijke foutmeldingen
- Juiste HTTP status codes

## Code Review Opmerkingen

### Configuratie

- Gebruik één centrale configuratie in `config/config.go`
- Vermijd dubbele configuratie in verschillende delen van de code
- Gebruik environment variabelen voor alle configuratie
- Waarschuw bij onveilige defaults in productie

### Database

- Gebruik indexen voor veelgebruikte velden
- Gebruik trigram indexen voor ILIKE zoekopdrachten
- Zorg dat `DB_DROP_TABLES` altijd `false` is in productie
- Gebruik migraties voor schema wijzigingen

### Middleware

- Consolideer middleware functies om duplicatie te voorkomen
- Gebruik consistente error responses
- Log alleen relevante acties

### Models

- Gebruik GORM tags voor betere database schema
- Implementeer BeforeSave hooks voor wachtwoord hashing
- Gebruik constanten voor entity types en action types

### Repositories

- Consolideer Delete methodes
- Verbeter error handling
- Gebruik prepared statements voor veiligheid

### API

- Implementeer consistente response structuur
- Valideer input parameters
- Beperk pageSize om database overbelasting te voorkomen
- Gebruik juiste HTTP status codes
