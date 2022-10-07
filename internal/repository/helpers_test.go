package repository

import (
	"os"
	"testing"

	"github.com/porter-dev/porter-agent/internal/adapter"
	"github.com/porter-dev/porter/api/server/shared/config/env"
	"gorm.io/gorm"
)

type tester struct {
	repo       *Repository
	dbFileName string
	db         *gorm.DB
}

func setupTestEnv(tester *tester, t *testing.T) {
	t.Helper()

	db, err := adapter.New(&env.DBConf{
		SQLLite:     true,
		SQLLitePath: tester.dbFileName,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	err = AutoMigrate(db, false)

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	tester.db = db
	tester.repo = NewRepository(db)
}

func cleanup(tester *tester, t *testing.T) {
	t.Helper()

	// remove the created file file
	os.Remove(tester.dbFileName)
}
