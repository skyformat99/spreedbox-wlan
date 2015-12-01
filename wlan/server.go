package wlan

import (
	"golang.struktur.de/spreedbox/spreedbox-go/bus"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	ec *bus.EncodedConn
}

func NewServer() (*Server, error) {
	s := &Server{}
	return s, nil
}

func (s *Server) Serve() (err error) {
	log.Println("connecting events")

	s.ec, err = bus.EstablishConnection(nil)
	if err != nil {
		return err
	}
	defer s.ec.Close()

	s.ec.Subscribe(WlanSubjectInterfaces(), s.interfaces)
	log.Println("events connected and subscribed")

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	return
}

func (s *Server) interfaces(subject, reply string, msg *InterfacesRequest) {
	log.Println("interfaces", subject, reply, msg.Names)

}
