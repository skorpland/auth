package test

import (
	"github.com/powerbase/auth/internal/conf"
	"github.com/powerbase/auth/internal/storage"
)

func SetupDBConnection(globalConfig *conf.GlobalConfiguration) (*storage.Connection, error) {
	return storage.Dial(globalConfig)
}
