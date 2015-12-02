package wlan

import (
	"encoding/json"
	"golang.struktur.de/spreedbox/spreedbox-conf/conf"
	"golang.struktur.de/spreedbox/spreedbox-go/bus"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan/linux"
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
	s.ec.Subscribe(WlanSubjectScan(), s.scan)
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

		iwgetid := linux.NewIWGetID()

		for _, i := range interfaces {
			if !network.IsInterfaceEthernet(i.Name) || !network.IsInterfaceWifi(i.Name) {
				continue
			}
			wi := &WlanInterface{
				ApAddress: iwgetid.Ap(i.Name),
				Frequency: iwgetid.Freq(i.Name),
				Channel:   iwgetid.Channel(i.Name),
				Protocol:  iwgetid.Protocol(i.Name),
				ESSID:     iwgetid.ESSID(i.Name),
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

	iwlist := linux.NewIWList()
	wlanCells, err := iwlist.Scan(msg.Name)
	if err != nil {
		log.Println("failed to scan", err)
	}

	pretty, _ := json.MarshalIndent(wlanCells, "", "\t")
	log.Println("scan result", string(pretty))

	if reply != "" {
		replyData, err := conf.NewDataReply(wlanCells, err)
		if err != nil {
			log.Println("failed to create interfaces reply", err)
			return
		}
		s.ec.Publish(reply, replyData)
	}
}
