package main

import (
	"flag"
	"fmt"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s: <command>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()
	command := args[0]
	switch command {
	case "defaultpassword":
		pw, err := wlan.GenerateDevicePassword(wlan.DefaultPasswordGeneratorVersion, wlan.DefaultPasswordLength)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stdout, "%s\n", pw)
	default:
		fmt.Fprintf(os.Stderr, "unknown command, %s\n", command)
		os.Exit(1)
	}
}
