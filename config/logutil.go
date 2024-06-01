package config

import (
	"bufio"
	"io"
	"log"
	"os"
)

var (
	logFile *os.File
	writer  *bufio.Writer
)

// SetupLogFile sets up logging to a file and to stdout.
func SetupLogFile(filename string, bufferSize int) {
	var err error
	logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a buffered writer
	writer = bufio.NewWriterSize(logFile, bufferSize)

	// Create a MultiWriter to write to both the buffered writer and stdout.
	multiWriter := io.MultiWriter(os.Stdout, writer)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Log file initialized")
}

// FlushLogFile flushes the buffered writer.
func FlushLogFile() {
	if writer != nil {
		err := writer.Flush()
		if err != nil {
			return
		}
	}
}

// CloseLogFile closes the log file if it is open.
func CloseLogFile() {
	FlushLogFile()
	if logFile != nil {
		logFile.Close()
	}
}
