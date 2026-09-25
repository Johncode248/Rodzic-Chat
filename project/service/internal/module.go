package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/reminders"
)

func Init() {
	background(context.Background())
}

func background(ctx context.Context) {
	fmt.Println("background start")

	go runReminderScheduler(ctx)
}

// runReminderScheduler is this backend's first scheduled/background job (see this file's git
// history -- background() used to do nothing else). It runs reminders.SendDueReminders once per
// calendar day, gated on an in-memory "already ran today" date rather than a persisted
// last-run timestamp -- a missed/duplicate run around a server restart is harmless here (see
// SendDueReminders' own doc comment for why it can never double-send within a single day
// anyway), so it wasn't worth a DB table just to survive restarts more precisely. Checking hourly
// rather than sleeping until a fixed clock time keeps this immune to the process's own start
// time landing before or after that hour.
func runReminderScheduler(ctx context.Context) {
	const reminderHourLocal = 8 // don't push before 8am local time
	lastRunDate := ""

	runIfDue := func() {
		now := time.Now()
		today := now.Format("2006-01-02")
		if today == lastRunDate || now.Hour() < reminderHourLocal {
			return
		}
		reminders.SendDueReminders(ctx, now)
		lastRunDate = today
	}

	runIfDue()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		runIfDue()
	}
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
