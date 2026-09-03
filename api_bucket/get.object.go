package api_bucket

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func GetObject(ctx context.Context, bucketName, objectName string) (res string, err error) {
	ctx = logger.Enter(ctx, "api_wasabi.GetObject", bucketName, objectName)
	defer logger.Exit(ctx, "api_wasabi.GetObject", &err)

	//obj, err := ClientBucket.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	//if err != nil {
	//	return "", err
	//}
	//defer obj.Close()
	//
	//data, err := io.ReadAll(obj)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(string(data))

	//reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//var text string
	//for i := 1; i <= reader.NumPage(); i++ {
	//	page := reader.Page(i)
	//	if page.V.IsNull() {
	//		continue
	//	}
	//
	//	content, err := page.GetPlainText(nil)
	//	if err != nil {
	//		return "", err
	//	}
	//
	//	text += content
	//}

	return "", nil
}
