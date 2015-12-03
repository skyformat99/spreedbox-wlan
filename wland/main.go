package main

import (
	"golang.struktur.de/spreedbox/spreedbox-go/common"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan"
	"log"
	"os"
)

func main() {
	if err := common.SetupLogging(); err != nil {
		log.Println("Could not setup logging:", err)
		os.Exit(1)
	}

	server, err := wlan.NewServer()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.Serve()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print("exiting")
}
