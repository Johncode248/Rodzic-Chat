package api_bucket

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func DeleteFile(ctx context.Context, bucketName, objectName string) (err error) {
	ctx = logger.Enter(ctx, "api_wasabi.DeleteFile", bucketName, objectName)
	defer logger.Exit(ctx, "api_wasabi.DeleteFile", &err)

	if err = ClientBucket.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{
		ForceDelete: true,
	}); err != nil {
		panic(err)
	}

	return nil
}
