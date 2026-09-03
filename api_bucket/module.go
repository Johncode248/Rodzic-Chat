package api_bucket

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ClientBucket *minio.Client

const (
	token     = "cfat_rVAYNdr8Ei6XwicBQgEXdCjDYpPckeS95x8komoQa6e97f95"
	accessKey = "1c5c1525a2e28d39ceb91f55a459af75"
	secretKey = "5c9e3d0cd3401be97737a6871c484fff79dde490528b40714a0a69968837e249"
	//url1      = "https://c39aac827c7f8f2e7e875ebab92c8f08.eu.r2.cloudflarestorage.com"
	endpoint = "c39aac827c7f8f2e7e875ebab92c8f08.r2.cloudflarestorage.com"
)

func Init() {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	ClientBucket = client

	fmt.Println("Created minio client")
}
