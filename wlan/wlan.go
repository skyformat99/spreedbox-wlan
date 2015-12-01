package wlan

import (
	"fmt"
)

const (
	BUS_WLAN_SUBJECT = "wlan"
)

type WlanSettings struct {
	Interfaces map[string]*WlanInterface `json:"devices,omitempty"`
}

type WlanInterface struct {
	ApMac     string `json:"apmac"`
	Frequency string `json:"frequency"`
	Channel   string `json:"channel"`
	Protocol  string `json:"protocol"`
	Essid     string `json:"essid"`
}

type InterfacesRequest struct {
	Names []string `json:"names,omitempty"`
}

func WlanSubjectInterfaces() string {
	return fmt.Sprintf("%s.interfaces", BUS_WLAN_SUBJECT)
}
