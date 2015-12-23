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

const (
	DISCOVER_WLAN_NAME = "wlan"
)

type Server struct {
	ec      *bus.EncodedConn
	scanner *Scanner
}

func NewServer() (*Server, error) {
	s := &Server{
		scanner: NewScanner(),
	}
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
	s.ec.Subscribe(WlanSubjectScan(), s.scan)
	s.ec.RegisterService(DISCOVER_WLAN_NAME)
	log.Println("events connected and subscribed")

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	return
}

func (s *Server) interfaces(subject, reply string, msg *InterfacesRequest) {
	log.Println("interfaces", subject, reply, msg.Names)
	go s.interfacesHandler(subject, reply, msg)
}

func (s *Server) interfacesHandler(subject, reply string, msg *InterfacesRequest) {
	wls := &WlanSettings{}

	// devices
	if interfaces, err := net.Interfaces(); err == nil {
		interfacesResult := make(map[string]*WlanInterface)

		interfaceGetter := NewWlanInterfaceGetter()

		for _, i := range interfaces {
			if !network.IsInterfaceEthernet(i.Name) || !network.IsInterfaceWifi(i.Name) {
				continue
			}
			wi, err := interfaceGetter.GetInterface(i.Name)
			if err != nil {
				log.Println("failed to get interface", i.Name, err)
				continue
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

func (s *Server) scan(subject, reply string, msg *ScanRequest) {
	log.Println("scan", subject, reply, msg.Name)
	go s.scanHandler(subject, reply, msg)
}

func (s *Server) scanHandler(subject, reply string, msg *ScanRequest) {
	wlanCells, err := s.scanner.Scan(msg.Name)

	if reply != "" {
		replyData, err := conf.NewDataReply(wlanCells, err)
		if err != nil {
			log.Println("failed to create interfaces reply", err)
			return
		}
		s.ec.Publish(reply, replyData)
	}
}
