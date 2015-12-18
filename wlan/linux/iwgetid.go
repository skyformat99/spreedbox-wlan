package linux

import (
	"log"
	"os/exec"
	"strings"
)

type IWGetID struct {
}

func NewIWGetID() *IWGetID {
	return &IWGetID{}
}

func (c *IWGetID) exec(interfaceName string, arg ...string) ([]byte, error) {
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

func (c *IWGetID) command(interfaceName, command string) string {
	out, err := c.exec(interfaceName, command)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *IWGetID) Ap(interfaceName string) string {
	return c.command(interfaceName, "--ap")
}

func (c *IWGetID) Freq(interfaceName string) string {
	result := c.command(interfaceName, "--freq")
	if strings.HasSuffix(result, "e+09") {
		result = result[:len(result)-4] + " GHz"
	}
	return result
}

func (c *IWGetID) Channel(interfaceName string) string {
	return c.command(interfaceName, "--channel")
}

func (c *IWGetID) Mode(interfaceName string) string {
	return c.command(interfaceName, "--mode")
}

func (c *IWGetID) Protocol(interfaceName string) string {
	return c.command(interfaceName, "--protocol")
}

func (c *IWGetID) ESSID(interfaceName string) string {
	return c.command(interfaceName, "")
}
