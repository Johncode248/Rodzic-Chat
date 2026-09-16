package api_bucket

import (
	"fmt"

	"github.com/Johncode248/Rodzic-Chat/variables"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ClientBucket *minio.Client

func Init() {
	client, err := minio.New(variables.R2Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(variables.R2AccessKey, variables.R2SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	ClientBucket = client

	fmt.Println("Created minio client")
}
