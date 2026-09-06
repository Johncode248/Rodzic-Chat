package api_bucket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

// downloadFilename is the name presented to the browser via
// Content-Disposition -- pass the original uploaded filename when known
// (e.g. "zgoda-wycieczka-zoo.pdf"), falling back to objectName otherwise.
func DownloadFile(ctx context.Context, bucketName, objectName, downloadFilename string) (res string, err error) {
	ctx = logger.Enter(ctx, "api_wasabi.DownloadFile", bucketName, objectName)
	defer logger.Exit(ctx, "api_wasabi.DownloadFile", &res, &err)

	if downloadFilename == "" {
		downloadFilename = objectName
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadFilename))

	result, err := ClientBucket.PresignedGetObject(ctx, bucketName, objectName, 1*time.Hour, reqParams)
	if err != nil {
		return "", err
	}

	res = fmt.Sprintf("https://%s%s?%s", result.Host, result.Path, result.RawQuery)

	return res, nil
}
