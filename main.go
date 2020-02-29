package main

import (
	"github.com/0990/momentscleaner/cleaner"
	"github.com/0990/momentscleaner/logconfig"
)

func main() {
	logconfig.InitLogrus("cleaner", 10)
	cleaner.DoClean()
}
