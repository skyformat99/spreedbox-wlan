package linux

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type IWList struct {
}

func NewIWList() *IWList {
	return &IWList{}
}

type IWListCell struct {
	Address             string                          `json:"address"`
	ESSID               string                          `json:"essid"`
	Mode                string                          `json:"mode"`
	Channel             string                          `json:"channel"`
	Frequency           string                          `json:"frequency"`
	EncryptionKeyStatus string                          `json:"encryptionKeyStatus"`
	InformationElements []*IWListCellInformationElement `json:"informationElements"`
	QualityLevel        float32                         `json:"qualityLevel"`
	SignalLevel         string                          `json:"signalLevel"`
	NoiseLevel          string                          `json:"noiseLevel"`
}

type IWListCellInformationElement struct {
	Protocol             string `json:"protocol"`
	GroupCipher          string `json:"groupcipher,omitempty"`
	PairwiseCiphers      string `json:"pairwiseciphers,omitempty"`
	AuthenticationSuites string `json:"authenticationSuites,omitempty"`
}

func (c *IWList) parse(data string) []*IWListCell {
	cells := []*IWListCell{}
	var cell *IWListCell
	var ie *IWListCellInformationElement

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
			cell = &IWListCell{
				InformationElements: []*IWListCellInformationElement{},
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
			ie = &IWListCellInformationElement{}
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
							cell.QualityLevel = float32(qualityLevel) / float32(qualityLevelMax)
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

func (c *IWList) Scan(interfaceName string) ([]*IWListCell, error) {
	cmd := exec.Command("sudo", "iwlist", interfaceName, "scan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("iwlist failed", interfaceName, err, string(out))
		return nil, err
	}
	cells := c.parse(string(out))
	return cells, err
}
