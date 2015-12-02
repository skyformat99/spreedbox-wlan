package wlan

import (
	"errors"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan/linux"
	"log"
	"sync"
)

type Scanner struct {
	sync.RWMutex
	wg       sync.WaitGroup
	scanning bool
	cells    []*linux.IWListCell
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(interfaceName string) (cells []*linux.IWListCell, err error) {
	if !network.IsInterfaceWifi(interfaceName) {
		return nil, errors.New("interface has no wifi extensions")
	}
	s.Lock()
	if !s.scanning {
		s.wg.Add(1)
		s.scanning = true
		s.Unlock()
		cells, err = s.scan(interfaceName)
		s.Lock()
		s.scanning = false
		if err == nil {
			s.cells = cells
		}
		s.wg.Done()
		s.Unlock()
	} else {
		s.Unlock()
		s.wg.Wait()
		s.RLock()
		cells = s.cells
		s.RUnlock()
	}
	return cells, err
}

func (s *Scanner) scan(interfaceName string) ([]*linux.IWListCell, error) {
	iwlist := linux.NewIWList()
	cells, err := iwlist.Scan(interfaceName)
	if err != nil {
		log.Println("failed to scan", err)
	}
	return cells, err
}
