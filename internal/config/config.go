package config

// Config bevat alle applicatie configuratie
type Config struct {
	JWTSecret          string
	JWTExpirationHours int
}
