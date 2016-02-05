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
		if os.Geteuid() != 0 {
			fmt.Fprintf(os.Stderr, "This command requires root permissions.\n")
			os.Exit(3)
		}
		pw, err := wlan.GenerateDevicePassword(wlan.DefaultPasswordGeneratorVersion, wlan.DefaultPasswordLength)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stdout, "%s\n", pw)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
