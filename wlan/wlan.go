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
	Protocol string `json:"protocol"`
	Ssid     string `json:"ssid"`
	Channel  string `json:"channel"`
	Quality  int8   `json:"quality"`
}

type InterfacesRequest struct {
	Names []string `json:"names,omitempty"`
}

func WlanSubjectInterfaces() string {
	return fmt.Sprintf("%s.interfaces", BUS_WLAN_SUBJECT)
}
