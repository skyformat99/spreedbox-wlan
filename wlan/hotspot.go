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
	runCmd        string
	restoreCmd    string
	deviceName    string
	passPhrase    string
	gracePeriod   time.Duration
	seenLinkMark  string
	seenLinkTimer *time.Timer
	link          bool
	started       bool
	running       bool
	quit          chan (bool)
	timer         *time.Timer
	cmd           *exec.Cmd
}

func NewHotspot(runCmd, restoreCmd, deviceName, passPhrase string, gracePeriod time.Duration, seenLinkMark string) *Hotspot {
	return &Hotspot{
		runCmd:       runCmd,
		restoreCmd:   restoreCmd,
		deviceName:   deviceName,
		passPhrase:   passPhrase,
		gracePeriod:  gracePeriod,
		seenLinkMark: seenLinkMark,
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
		// Link and running.
		if h.cmd != nil && len(deviceNames) == 1 && deviceNames[0] == h.deviceName {
			// It is our device and the have a running cmd, do nothing.
			log.Println("hotspot is running with link on own device", h.deviceName)
			return
		}
		// Not our deivce.
		h.markSeenLink()
		log.Println("stopping hotspot as there is a link on other device", deviceNames)
		h.stop(true)
	} else if link {
		// Link but not running, other device has link.
		h.markSeenLink()
	}
	if !link {
		if h.seenLinkTimer != nil {
			// Kill seen link which might be in progress.
			h.seenLinkTimer.Stop()
			h.seenLinkTimer = nil
		}
		if !h.running {
			// No link and not running.
			h.start()
		} else if h.cmd != nil && h.cmd.Process != nil {
			// No link and have cmd, trigger exit - probably device removed.
			log.Println("terminating hotspot as no link and running - device removed?")
			h.cmd.Process.Signal(syscall.SIGTERM)
		}
	}
}

func (h *Hotspot) Exit() {
	h.Lock()
	defer h.Unlock()
	h.started = false
	h.stop(true)
}

func (h *Hotspot) Reset() {
	h.Lock()
	defer h.Unlock()
	// Unmark link status, reenable hotspot on reset to avoid reboot required
	// a non-working network configuration is applied. Reset() is called when
	// spreedbox-network ifdowns the network.
	h.unmarkSeenLink()
	if h.running {
		log.Println("hotspot reset requested")
		h.stop(false)
		h.start()
	}
}

func (h *Hotspot) markSeenLink() {
	if h.seenLinkMark == "" {
		return
	}

	if h.seenLinkTimer != nil {
		h.seenLinkTimer.Stop()
	}

	// Mark after we are certain.
	h.seenLinkTimer = time.AfterFunc(30*time.Second, h.doMarkSeenLink)
}

func (h *Hotspot) doMarkSeenLink() {
	h.Lock()
	defer h.Unlock()

	if h.seenLinkTimer == nil {
		// Cleaned up already.
		return
	}
	h.seenLinkTimer = nil
	if !h.link {
		// No more link, ignore mark.
		return
	}

	if _, err := os.Stat(h.seenLinkMark); os.IsNotExist(err) {
		// Create mark.
		err := ioutil.WriteFile(h.seenLinkMark, []byte{}, 644)
		if err != nil {
			log.Println("failed to write link seen mark", err)
		} else {
			log.Println("set link seen mark, automatic hotspot is now disabled")
		}
	}
}

func (h *Hotspot) unmarkSeenLink() {
	if h.seenLinkMark == "" {
		return
	}

	err := os.Remove(h.seenLinkMark)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Println("failed to remove write link seen mark", err)
	} else {
		log.Println("unset link seen mark, automatic hotspot is now enabled")
	}
}

func (h *Hotspot) hasSeenLink() bool {
	if h.seenLinkMark == "" {
		return false
	}

	_, err := os.Stat(h.seenLinkMark)
	if os.IsNotExist(err) {
		return false
	} else if err == nil {
		return true
	}

	return false
}

func (h *Hotspot) stop(restore bool) {
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
	} else {
		// No need to restore, if was not waiting on quit.
		restore = false
	}

	if restore && h.restoreCmd != "" {
		log.Println("restoring device after hotspot exit ...")
		command := strings.Split(h.restoreCmd, " ")
		command = append(command, h.deviceName)
		cmd := exec.Command(command[0], command[1:]...)
		if err := cmd.Start(); err != nil {
			log.Println("failed to restore device after hotspot exit", err)
		} else {
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case <-time.After(10 * time.Second):
				log.Println("restore device timed out")
				if err := cmd.Process.Kill(); err != nil {
					log.Println("failed to kill restore device process", err)
				}
			case err := <-done:
				if err != nil {
					log.Println("restore device after hotspot failed", err)
				}
			}
		}
	}
}

func (h *Hotspot) start() {
	if h.deviceName == "" || h.runCmd == "" {
		return
	}
	if h.hasSeenLink() {
		log.Println("hotspot is disabled (link seen)")
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
		h.stop(false)
		h.Unlock()
	}

}
