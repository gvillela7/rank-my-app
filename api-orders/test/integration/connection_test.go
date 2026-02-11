package integration

import (
	"testing"

	"github.com/gvillela7/rank-my-app/configs"
)

func TestMongoDBConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Carregar config do diretório raiz do projeto
	if err := config.Load("../.."); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verificar configuração
	mongoCfg := config.GetDBMongo()
	t.Logf("MongoDB URI: %s", mongoCfg.URI)
	t.Logf("MongoDB Database: %s", mongoCfg.Database)

	if mongoCfg.URI == "" {
		t.Fatal("MongoDB URI is empty - check config.toml")
	}

	t.Log("✅ Config loaded successfully!")
}
