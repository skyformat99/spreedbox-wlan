package wlan

import (
	"bufio"
	"fmt"
	"golang.struktur.de/spreedbox/spreedbox-network/network"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Hotspot struct {
	sync.RWMutex
	runCmd      string
	deviceName  string
	passPhrase  string
	gracePeriod time.Duration
	link        bool
	started     bool
	running     bool
	quit        chan (bool)
	timer       *time.Timer
	cmd         *exec.Cmd
}

func NewHotspot(runCmd, deviceName, passPhrase string, gracePeriod time.Duration) *Hotspot {
	return &Hotspot{
		runCmd:      runCmd,
		deviceName:  deviceName,
		passPhrase:  passPhrase,
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
	}

	if link && h.running {
		if len(deviceNames) == 1 && deviceNames[0] == h.deviceName {
			// It is our device, do nothing.
			log.Println("hotspot is running with link on own device", h.deviceName)
			return
		}
		log.Println("stopping hotspot as there is a link on other device", deviceNames)
		h.stop()
	}
	if !link && !h.running {
		h.start()
	}
}

func (h *Hotspot) Exit() {
	h.Lock()
	defer h.Unlock()
	h.started = false
	h.stop()
}

func (h *Hotspot) stop() {
	if h.running {
		log.Println("hotspot stop")
		h.running = false
	}
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	if h.cmd != nil {
		if h.cmd.Process != nil {
			h.cmd.Process.Signal(syscall.SIGTERM)
		}
		log.Println("waiting for hotspot to exit ...")
		h.cmd.Wait()
		h.cmd = nil
	}

	if h.quit != nil {
		<-h.quit
		h.quit = nil
	}
}

func (h *Hotspot) start() {
	if h.deviceName == "" || h.runCmd == "" {
		return
	}
	log.Println("hotspot start scheduled in", h.gracePeriod)
	h.running = true
	h.timer = time.AfterFunc(h.gracePeriod, h.run)
}

func (h *Hotspot) makePskFile() (string, error) {
	if h.passPhrase == "" {
		return "", nil
	}
	pskFile, err := ioutil.TempFile(os.TempDir(), "wland")
	if err != nil {
		return "", err
	}
	defer pskFile.Close()
	if _, err := pskFile.WriteString(fmt.Sprintf("00:00:00:00:00:00 %s\n", h.passPhrase)); err != nil {
		return "", err
	}
	return pskFile.Name(), nil
}

func (h *Hotspot) run() {
	h.Lock()
	if !h.running {
		h.Unlock()
		return
	}

	if !network.IsInterfaceWifi(h.deviceName) {
		// Do nothing, when device is not there or no wifi device
		h.timer = time.AfterFunc(h.gracePeriod, h.run)
		h.Unlock()
		return
	}

	// Prepare command
	command := strings.Split(h.runCmd, " ")
	command = append(command, h.deviceName)

	pskFileName, err := h.makePskFile()
	if err != nil {
		log.Println("failed to create PSK file", err)
		return
	}
	if pskFileName != "" {
		command = append(command, pskFileName)
		log.Println("hotspot will be encrypted")
	} else {
		log.Println("hotspot will not be encrypted")
	}

	log.Println("starting hotspot ...")
	cmd := exec.Command(command[0], command[1:]...)
	h.cmd = cmd
	h.quit = make(chan bool, 1)
	h.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("failed to run hotspot command", err)
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
	if pskFileName != "" {
		os.Remove(pskFileName)
	}
	h.quit <- true

	h.Lock()
	if h.running && h.cmd == cmd {
		err = cmd.Wait()
		log.Println("hotspot unexpected exit", err)
		h.Unlock()
		// Restart when still marked running.
		h.run()
	} else {
		h.stop()
		h.Unlock()
	}

}
