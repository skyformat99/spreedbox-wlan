package wlan

import (
	"bufio"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Hotspot struct {
	sync.RWMutex
	runCmd      string
	deviceName  string
	gracePeriod time.Duration
	link        bool
	started     bool
	running     bool
	timer       *time.Timer
	cmd         *exec.Cmd
}

func NewHotspot(runCmd, deviceName string, gracePeriod time.Duration) *Hotspot {
	return &Hotspot{
		runCmd:      runCmd,
		deviceName:  deviceName,
		gracePeriod: gracePeriod,
	}
}

func (h *Hotspot) SetLink(link bool, deviceNames []string) {
	h.Lock()
	defer h.Unlock()
	if !h.started || h.link != link {
		h.started = true
		h.link = link
		log.Println("link status changed", link)

		if link && h.running {
			if len(deviceNames) == 1 && deviceNames[0] == h.deviceName {
				// It is our device, do nothing.
				log.Println("hotspot is running with link on own device", h.deviceName)
				return
			}
			h.stop()
		}
		if !link && !h.running {
			h.start()
		}
	}
}

func (h *Hotspot) stop() {
	log.Println("hotspot stop")
	h.running = false
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	if h.cmd != nil {
		if h.cmd.Process != nil {
			h.cmd.Process.Kill()
		}
		h.cmd = nil
	}
}

func (h *Hotspot) start() {
	if h.deviceName == "" || h.runCmd == "" {
		return
	}
	log.Println("hotspot start")
	h.running = true
	h.timer = time.AfterFunc(h.gracePeriod, h.run)
}

func (h *Hotspot) run() {
	h.Lock()
	if !h.running {
		h.Unlock()
		return
	}

	if !network.IsInterfaceWifi(h.deviceName) {
		// Do nothing, when device is not there or no wifi device.
		h.timer = time.AfterFunc(h.gracePeriod, h.run)
		h.Unlock()
		return
	}

	log.Println("starting hotspot")
	command := strings.Split(h.runCmd, " ")
	command = append(command, h.deviceName)
	cmd := exec.Command(command[0], command[1:]...)
	h.cmd = cmd
	h.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("failed to run hotspot", err)
		return
	}
	wait := make(chan bool, 1)
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
		wait <- true
	}()

	if err := cmd.Start(); err != nil {
		log.Println("failed to start hotspot", err)
	}

	<-wait
	err = cmd.Wait()
	log.Println("hotspot done", err)

	h.Lock()
	if h.running && h.cmd == cmd {
		h.Unlock()
		// Restart when still marked running.
		h.run()
	} else {
		h.stop()
		h.Unlock()
	}
}
