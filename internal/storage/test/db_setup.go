package test

import (
	"github.com/skorpland/auth/internal/conf"
	"github.com/skorpland/auth/internal/storage"
)

func SetupDBConnection(globalConfig *conf.GlobalConfiguration) (*storage.Connection, error) {
	return storage.Dial(globalConfig)
}
