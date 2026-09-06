package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
)

func Init() {
	background(context.Background())
}

func background(ctx context.Context) {
	fmt.Println("background start")

}

func upload() {
	filePath := "test2.txt"

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("nie udało się otworzyć pliku: %v", err)
	}
	defer file.Close()

	// Pobranie rozmiaru pliku
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("nie udało się pobrać info o pliku: %v", err)
	}
	size := fileInfo.Size()

	err = api_bucket.UploadFile(context.Background(), "dom2sky", "test3.txt", file, size, "")
	if err != nil {
		log.Fatalf("błąd uploadu: %v", err)
	}

	fmt.Println("Plik przesłany pomyślnie")
}

func download() {
	object, err := api_bucket.DownloadFile(context.Background(), "dom2sky", "zlecenie-1/files5.08.zip", "")
	if err != nil {
		panic(err)
	}

	fmt.Println(object)
}
