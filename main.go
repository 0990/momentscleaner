package main

import (
	"github.com/0990/momentscleaner/cleaner"
	"github.com/0990/momentscleaner/logconfig"
	_ "net/http/pprof"
)

func main() {
	//go func() {
	//	fmt.Println(http.ListenAndServe("localhost:8888", nil))
	//}()

	logconfig.InitLogrus("cleaner", 10)
	cleaner.DoClean()
}
