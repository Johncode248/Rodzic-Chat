package api_bucket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func DownloadFile(ctx context.Context, bucketName, objectName string) (res string, err error) {
	ctx = logger.Enter(ctx, "api_wasabi.DownloadFile", bucketName, objectName)
	defer logger.Exit(ctx, "api_wasabi.DownloadFile", &res, &err)

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, objectName))

	result, err := ClientBucket.PresignedGetObject(ctx, bucketName, objectName, 1*time.Hour, reqParams)
	if err != nil {
		return "", err
	}

	res = fmt.Sprintf("https://%s%s?%s", result.Host, result.Path, result.RawQuery)

	return res, nil
}
