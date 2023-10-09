package goearth

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ExtArgs struct {
	Port     int
	Cookie   string
	Filename string
}

var args ExtArgs
var flagsRegistered = false

// Registers the port, cookie, and filename command-line arguments with the flag package.
func RegisterFlags() *ExtArgs {
	if !flagsRegistered {
		flag.IntVar(&args.Port, "p", 9092, "The port to connect to G-Earth on.")
		flag.StringVar(&args.Cookie, "c", "", "The cookie provided by G-Earth.")
		flag.StringVar(&args.Filename, "f", "", "The filename provided by G-Earth.")
		flagsRegistered = true
	}
	return &args
}

func parseArgs() *ExtArgs {
	if flagsRegistered {
		flag.Parse()
		return &args
	}
	args := &ExtArgs{Port: 9092}
	for i := 0; i < len(os.Args)-1; i++ {
		next := os.Args[i+1]
		switch os.Args[i] {
		case "-p":
			port, err := strconv.Atoi(next)
			if err != nil {
				panic(fmt.Errorf("failed to parse port: %q", next))
			}
			args.Port = port
		case "-c":
			args.Cookie = next
		case "-f":
			args.Filename = next
		default:
			continue
		}
		i++
	}
	return args
}
