package linux

import (
	"log"
	"os/exec"
	"strings"
)

type IWList struct {
}

func NewIWList() *IWList {
	return &IWList{}
}

type IWListCell struct {
	Address             string                         `json:"address"`
	ESSID               string                         `json:"essid"`
	Protocol            string                         `json:"protocol"`
	Mode                string                         `json:"mode"`
	Frequency           string                         `json:"frequency"`
	EncryptionKeyStatus string                         `json:"encryptionKeyStatus"`
	BitRates            string                         `json:"bitrates"`
	InformationElements []IWListCellInformationElement `json:"informationElements"`
	QualityLevel        int                            `json:"qualityLevel"`
	SignalLevel         int                            `json:"signalLevel"`
	NoiseLevel          int                            `json:"noiseLevel"`
}

type IWListCellInformationElement struct {
	Protocol             string `json:"protocol"`
	GroupCipher          string `json:"groupcipher"`
	PairwiseCiphers      string `json:"pairwiseciphers"`
	AuthenticationSuites string `json:"authenticationSuites"`
	Extra                string `json:"extra"`
}

func (c *IWList) parse(data string) []IWListCell {
	cells := []IWListCell{}
	/*cell := &IWListCell{}
	ies := []IWListCellInformationElement{}
	ie := &IWListCellInformationElement{}*/

	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		log.Println("scan line", line)

	}
	return cells
}

func (c *IWList) Scan(interfaceName string) ([]IWListCell, error) {
	cmd := exec.Command("iwlist", interfaceName, "scan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("iwlist failed", interfaceName, err, string(out))
		return nil, err
	}
	cells := c.parse(string(out))
	return cells, err
}
