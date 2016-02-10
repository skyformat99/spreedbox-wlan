package wlan

import (
	"golang.struktur.de/spreedbox/spreedbox-conf/conf"
	"golang.struktur.de/spreedbox/spreedbox-go/bus"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	DISCOVER_WLAN_NAME = "wlan"
)

var DefaultLinkCheckTimeout = 5 * time.Second
var DefaultHotspotCommand = ""
var DefaultHotspotInterface = "wlan0"
var DefaultHotspotGracePeriod = 60 * time.Second
var DefaultHotspotPassPhrase = "spreedbox"

var setupServerOnce sync.Once

type Server struct {
	ec      *bus.EncodedConn
	scanner *Scanner
	hotspot *Hotspot
}

func NewServer() (*Server, error) {
	setupServerOnce.Do(setupServer)
	s := &Server{
		scanner: NewScanner(),
		hotspot: NewHotspot(
			DefaultHotspotCommand,
			DefaultHotspotInterface,
			DefaultHotspotPassPhrase,
			DefaultHotspotGracePeriod,
		),
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
	s.ec.Subscribe((&network.LinkChangedEvent{}).Subject(), func(event *network.LinkChangedEvent) {
		s.hotspot.SetLink(event.Link, event.DeviceNames)
	})
	s.ec.Subscribe((&network.ApplyStoppingEvent{}).Subject(), func(event *network.ApplyStoppingEvent) {
		s.hotspot.Reset()
	})
	s.ec.RegisterService(DISCOVER_WLAN_NAME)
	log.Println("events connected and subscribed")

	go func() {
		link := &network.LinkReply{}
		s.ec.Request(network.NetworkSubjectLink(), &network.LinkRequest{}, &link, DefaultLinkCheckTimeout)
		if link.Success {
			s.hotspot.SetLink(link.Link, []string{})
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	s.hotspot.Exit()
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
	wlanCells, err := s.scanner.Scan(msg.Name, msg.Rescan)

	if reply != "" {
		replyData, err := conf.NewDataReply(wlanCells, err)
		if err != nil {
			log.Println("failed to create interfaces reply", err)
			return
		}
		s.ec.Publish(reply, replyData)
	}
}

func setupServer() {
	hotspotCommand := os.Getenv("HOTSPOT_COMMAND")
	if hotspotCommand != "" {
		DefaultHotspotCommand = hotspotCommand
	}
	hotspotInterface := os.Getenv("HOTSPOT_INTERFACE")
	if hotspotInterface != "" {
		DefaultHotspotInterface = hotspotInterface
	}
	hotspotGracePeriod := os.Getenv("HOTSPOT_GRACEPERIOD")
	if hotspotGracePeriod != "" {
		if hotspotGracePeriodInt, err := strconv.Atoi(hotspotGracePeriod); err == nil {
			DefaultHotspotGracePeriod = time.Duration(hotspotGracePeriodInt) * time.Second
		}
	}
	if os.Getenv("HOTSPOT_UNENCRYPTED") != "" {
		DefaultHotspotPassPhrase = ""
	} else {
		hotspotPassPhrase := os.Getenv("HOTSPOT_PASSPHRASE")
		if hotspotPassPhrase != "" {
			DefaultHotspotPassPhrase = hotspotPassPhrase
		} else {
			// Use hardware specific password by default.
			if hwSpecificPassPhrase, err := GenerateDevicePassword(DefaultPasswordGeneratorVersion, DefaultPasswordLength); err == nil {
				log.Printf("using hardware based default hotspot password (version %d)\n", DefaultPasswordGeneratorVersion)
				DefaultHotspotPassPhrase = hwSpecificPassPhrase
			} else {
				log.Println("failed to get hardware based default hotspot password:", err)
			}
		}
	}
}
