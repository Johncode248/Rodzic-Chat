package main

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

func BigMigrate() {

	if err := gormdb.Migrate(context.Background(), credentialspg.Credential{}); err != nil {
		panic(err)
	}

}

/*
DROP TABLE agencies;
DROP TABLE credentials;
DROP TABLE dron_media;
DROP TABLE drone_requests;
DROP TABLE media;
DROP TABLE reports;
DROP TABLE users;


*/
