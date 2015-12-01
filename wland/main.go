package main

import (
	"golang.struktur.de/spreedbox/spreedbox-go/common"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan"
	"log"
)

func main() {
	common.SetupLogging()

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
