package wlan

import (
	"errors"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan/linux"
	"log"
	"sync"
	"sync/atomic"
)

// interfaceScanner scans a interface exactly once
type interfaceScanner struct {
	refCount int32

	interfaceName string
	cells         []*linux.IWListCell
	scanError     error

	sync.Once
}

func (s *interfaceScanner) scan() ([]*linux.IWListCell, error) {
	s.Do(func() {
		iwlist := linux.NewIWList()
		cells, err := iwlist.Scan(s.interfaceName)
		if err != nil {
			log.Println("failed to scan", err)
		}
		s.cells = cells
		s.scanError = err
	})

	return s.cells, s.scanError
}

type Scanner struct {
	sync.Mutex
	scanners map[string]*interfaceScanner
}

func NewScanner() *Scanner {
	return &Scanner{
		scanners: make(map[string]*interfaceScanner),
	}
}

func (s *Scanner) Scan(interfaceName string) (cells []*linux.IWListCell, err error) {
	if !network.IsInterfaceWifi(interfaceName) {
		// NOTE: spreedbox-setup check for exactly this message to generate
		// a proper error response.
		return nil, errors.New("interface has no wifi extensions")
	}

	s.Lock()
	scanner, found := s.scanners[interfaceName]
	if !found {
		scanner = &interfaceScanner{
			interfaceName: interfaceName,
		}
		s.scanners[interfaceName] = scanner
	}
	s.Unlock()

	atomic.AddInt32(&scanner.refCount, 1)
	cells, err = scanner.scan()

	if atomic.AddInt32(&scanner.refCount, -1) == 0 {
		s.Lock()
		delete(s.scanners, interfaceName)
		s.Unlock()
	}

	return cells, err
}
