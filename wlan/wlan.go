package wlan

import (
	"fmt"
)

const (
	BUS_WLAN_SUBJECT = "wlan"
)

type InterfacesRequest struct {
	Names []string `json:"names,omitempty"`
}

func WlanSubjectInterfaces() string {
	return fmt.Sprintf("%s.interfaces", BUS_WLAN_SUBJECT)
}
