// +build linux

package wlan

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

var (
	iwlistCmd   = "/sbin/iwlist"
	ifconfigCmd = "/sbin/ifconfig"
)

type LinuxWlanInterfaceScanner struct {
}

func NewWlanInterfaceScanner() *LinuxWlanInterfaceScanner {
	return &LinuxWlanInterfaceScanner{}
}

func (c *LinuxWlanInterfaceScanner) parse(data string) []*WlanInterfaceCell {
	cells := []*WlanInterfaceCell{}
	var cell *WlanInterfaceCell
	var ie *WlanInterfaceCellInformationElement

	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		splitter := ":"
		if !strings.Contains(line, ":") {
			splitter = "="
		}
		cols := strings.SplitN(line, splitter, 2)
		if len(cols) < 2 {
			continue
		}
		firstCol := strings.TrimSpace(cols[0])
		secondCol := strings.TrimSpace(cols[1])

		// Parse
		switch {
		case strings.HasPrefix(firstCol, "Cell"):
			// Cell 05 - Address: XX:XX:XX:XX:XX:XX lines trigger a new cell
			cell = &WlanInterfaceCell{
				InformationElements: []*WlanInterfaceCellInformationElement{},
			}
			cells = append(cells, cell)
			cell.Address = secondCol
		case firstCol == "ESSID":
			cell.ESSID = strings.Trim(secondCol, "\"")
		case firstCol == "Channel":
			cell.Channel = secondCol
		case firstCol == "Mode":
			cell.Mode = secondCol
		case firstCol == "Frequency":
			cell.Frequency = secondCol
		case firstCol == "Encryption key":
			cell.EncryptionKeyStatus = secondCol
		case firstCol == "IE":
			ie = &WlanInterfaceCellInformationElement{}
			cell.InformationElements = append(cell.InformationElements, ie)
			ie.Protocol = secondCol
		case strings.HasPrefix(firstCol, "Group Cipher"):
			ie.GroupCipher = secondCol
		case strings.HasPrefix(firstCol, "Pairwise Ciphers"):
			ie.PairwiseCiphers = secondCol
		case strings.HasPrefix(firstCol, "Authentication Suites"):
			ie.AuthenticationSuites = secondCol
		case firstCol == "Quality":
			// Quality needs own splitting
			qualityCols := strings.Split(line, "  ")
			for _, qualityCol := range qualityCols {
				qualityCol = strings.TrimSpace(qualityCol)
				qualitySplitter := ":"
				if !strings.Contains(line, ":") {
					qualitySplitter = "="
				}
				signalCols := strings.SplitN(qualityCol, qualitySplitter, 2)
				if len(signalCols) < 2 {
					continue
				}
				signalFirstCol := strings.TrimSpace(signalCols[0])
				signalSecondCol := strings.TrimSpace(signalCols[1])
				switch signalFirstCol {
				case "Quality":
					qualityLevelCols := strings.SplitN(signalSecondCol, "/", 2)
					if len(qualityLevelCols) < 2 {
						continue
					}
					qualityLevel, _ := strconv.Atoi(qualityLevelCols[0])
					if qualityLevelMax, err := strconv.Atoi(qualityLevelCols[1]); err == nil {
						if qualityLevelMax > 0 {
							cell.QualityLevel = int8(float32(qualityLevel) / float32(qualityLevelMax) * 100)
						}
					}
				case "Signal level":
					cell.SignalLevel = signalSecondCol
				case "Noise level":
					cell.NoiseLevel = signalSecondCol
				}
			}
		}
	}
	return cells
}

func (c *LinuxWlanInterfaceScanner) upInterface(name string) error {
	args := []string{name, "up"}
	cmd := exec.Command(ifconfigCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("ifconfig failed", name, err, string(out))
	}
	return err
}

func (c *LinuxWlanInterfaceScanner) ScanInterface(name string, rescan bool) ([]*WlanInterfaceCell, error) {
	c.upInterface(name) // Linux wifi interfaces need to be up to get scan results.

	args := []string{name, "scan"}
	if !rescan {
		args = append(args, "last")
	}
	log.Println("scanning", name, rescan)
	cmd := exec.Command(iwlistCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("iwlist failed", name, err, string(out))
		return nil, err
	}
	log.Println("scanning complete", name)
	cells := c.parse(string(out))
	return cells, err
}
