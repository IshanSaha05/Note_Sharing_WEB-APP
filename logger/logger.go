package logger

import (
	"log"
	"os"
)

var Log *log.Logger

func init() {
	file, err := os.OpenFile("C:\\Users\\User\\Desktop\\GoLang\\Project\\app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatalf("Error: Problems while opening log file.\n\tError: %s", err)
	}

	Log = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}
