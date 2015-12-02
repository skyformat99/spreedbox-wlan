package wlan

import (
	"fmt"
	"golang.struktur.de/spreedbox/spreedbox-wlan/wlan/linux"
)

const (
	BUS_WLAN_SUBJECT = "wlan"
)

type WlanSettings struct {
	Interfaces map[string]*WlanInterface `json:"devices,omitempty"`
}

type WlanInterface struct {
	ApAddress string `json:"apAddress"`
	Frequency string `json:"frequency"`
	Channel   string `json:"channel"`
	Protocol  string `json:"protocol"`
	ESSID     string `json:"essid"`
}

type InterfacesRequest struct {
	Names []string `json:"names,omitempty"`
}

type WlanCell struct {
	linux.IWListCell
	InformationElements []WlanCellInformationElement `json:"informationElements"`
}

type WlanCellInformationElement struct {
	linux.IWListCellInformationElement
}

type ScanRequest struct {
	Name string `json:"name"`
}

func WlanSubjectInterfaces() string {
	return fmt.Sprintf("%s.interfaces", BUS_WLAN_SUBJECT)
}

func WlanSubjectScan() string {
	return fmt.Sprintf("%s.scan", BUS_WLAN_SUBJECT)
}
