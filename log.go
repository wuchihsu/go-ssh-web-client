package main

import (
	"fmt"
	"log"
)

// func to print warn log with color yellow
func logWarn(msg string, args ...interface{}) {
	log.Printf("\033[33mwarn: %s\033[0m", fmt.Sprintf(msg, args...))
}

// func to print info log with color green
func logInfo(msg string, args ...interface{}) {
	log.Printf("\033[32minfo: %s\033[0m", fmt.Sprintf(msg, args...))
}

// func to print debug log with color blue
func logDebug(msg string, args ...interface{}) {
	log.Printf("\033[34mdebug: %s\033[0m", fmt.Sprintf(msg, args...))
}

// func to print error log with color red
func logError(msg string, args ...interface{}) {
	log.Printf("\033[31merror: %s\033[0m", fmt.Sprintf(msg, args...))
}
