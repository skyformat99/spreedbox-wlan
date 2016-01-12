package wlan

import (
	"errors"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"log"
	"sync"
	"sync/atomic"
)

// interfaceScanner scans a interface exactly once
type interfaceScanner struct {
	refCount int32

	interfaceName string
	rescan        bool
	cells         []*WlanInterfaceCell
	scanError     error

	sync.Once
}

func (s *interfaceScanner) scan() ([]*WlanInterfaceCell, error) {
	s.Do(func() {
		scanner := NewWlanInterfaceScanner()
		cells, err := scanner.ScanInterface(s.interfaceName, s.rescan)
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

func (s *Scanner) Scan(interfaceName string, rescan bool) (cells []*WlanInterfaceCell, err error) {
	if !network.IsInterfaceWifi(interfaceName) {
		// NOTE: spreedbox-setup check for exactly this message to generate
		// a proper error response.
		return nil, errors.New("interface has no wifi extensions")
	}

	s.Lock()
	scanner, found := s.scanners[interfaceName]
	if !found || (rescan && !scanner.rescan) {
		// First scan or a rescan request with the currently active scan is returning a cached list.
		scanner = &interfaceScanner{
			interfaceName: interfaceName,
			rescan:        rescan,
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
