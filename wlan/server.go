package wlan

import (
	"golang.struktur.de/spreedbox/spreedbox-conf/conf"
	"golang.struktur.de/spreedbox/spreedbox-go/bus"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"log"
	"net"
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
	wls := &WlanSettings{}

	// devices
	if interfaces, err := net.Interfaces(); err == nil {
		interfacesResult := make(map[string]*WlanInterface)

		for _, i := range interfaces {
			if !network.IsInterfaceEthernet(i.Name) || !network.IsInterfaceWifi(i.Name) {
				continue
			}
			wi := &WlanInterface{
			// TODO(longsleep): Add details
			}
			interfacesResult[i.Name] = wi
		}
		wls.Interfaces = interfacesResult

	} else {
		log.Println("failed to read interfaces", err)
	}

	if reply != "" {
		replyData, err := conf.NewDataReply(wls, nil)
		if err != nil {
			log.Println("failed to create interfaces reply", err)
			return
		}
		s.ec.Publish(reply, replyData)
	}
}
