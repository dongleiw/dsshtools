package main

import (
	"fmt"
	"log"
	"os"
)

var (
	g_debug = false
)

func enable_debug() {
	g_debug = true
	log.SetPrefix("DEBUG ")
	log.SetFlags(0x37)
}
func logdebug(fmtstr string, args ...interface{}) {
	if !g_debug {
		return
	}
	log.Printf(fmtstr, args...)
}
func logerr(fmtstr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtstr+"\n", args...)
}
