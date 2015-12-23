// +build linux

package wlan

import (
	"log"
	"os/exec"
	"strings"
)

type LinuxWlanInterfaceGetter struct {
}

func NewWlanInterfaceGetter() *LinuxWlanInterfaceGetter {
	return &LinuxWlanInterfaceGetter{}
}

func (c *LinuxWlanInterfaceGetter) exec(interfaceName string, arg ...string) ([]byte, error) {
	arguments := []string{interfaceName, "--raw"}
	if len(arg) > 0 {
		if len(arg) == 1 && arg[0] == "" {
		} else {
			arguments = append(arguments, arg...)
		}
	}
	cmd := exec.Command("iwgetid", arguments...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("iwgetid failed", interfaceName, arg, err, string(out))
		return nil, err
	}
	return out, err
}

func (c *LinuxWlanInterfaceGetter) command(interfaceName, command string) string {
	out, err := c.exec(interfaceName, command)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *LinuxWlanInterfaceGetter) Ap(interfaceName string) string {
	return c.command(interfaceName, "--ap")
}

func (c *LinuxWlanInterfaceGetter) Freq(interfaceName string) string {
	result := c.command(interfaceName, "--freq")
	if strings.HasSuffix(result, "e+09") {
		result = result[:len(result)-4] + " GHz"
	}
	return result
}

func (c *LinuxWlanInterfaceGetter) Channel(interfaceName string) string {
	return c.command(interfaceName, "--channel")
}

func (c *LinuxWlanInterfaceGetter) Mode(interfaceName string) string {
	return c.command(interfaceName, "--mode")
}

func (c *LinuxWlanInterfaceGetter) Protocol(interfaceName string) string {
	return c.command(interfaceName, "--protocol")
}

func (c *LinuxWlanInterfaceGetter) ESSID(interfaceName string) string {
	return c.command(interfaceName, "")
}

func (c *LinuxWlanInterfaceGetter) GetInterface(name string) (*WlanInterface, error) {
	wi := &WlanInterface{
		ApAddress: c.Ap(name),
		Frequency: c.Freq(name),
		Channel:   c.Channel(name),
		Protocol:  c.Protocol(name),
		ESSID:     c.ESSID(name),
	}
	return wi, nil
}
