package wlan

type WlanInterfaceCell struct {
	Address             string                                 `json:"address"`
	ESSID               string                                 `json:"essid"`
	Mode                string                                 `json:"mode"`
	Channel             string                                 `json:"channel"`
	Frequency           string                                 `json:"frequency"`
	EncryptionKeyStatus string                                 `json:"encryptionKeyStatus"`
	InformationElements []*WlanInterfaceCellInformationElement `json:"informationElements"`
	QualityLevel        int8                                   `json:"qualityLevel"`
	SignalLevel         string                                 `json:"signalLevel"`
	NoiseLevel          string                                 `json:"noiseLevel"`
}

type WlanInterfaceCellInformationElement struct {
	Protocol             string `json:"protocol"`
	GroupCipher          string `json:"groupCipher,omitempty"`
	PairwiseCiphers      string `json:"pairwiseCiphers,omitempty"`
	AuthenticationSuites string `json:"authenticationSuites,omitempty"`
}

type WlanInterfaceGetter interface {
	GetInterface(name string) (*WlanInterface, error)
}

// func NewWlanInterfaceGetter() WlanInterfaceGetter
// will be provided by OS-dependent implementations

type WlanInterfaceScanner interface {
	ScanInterface(name string, rescan bool) ([]*WlanInterfaceCell, error)
}

// func NewWlanInterfaceScanner() WlanInterfaceScanner
// will be provided by OS-dependent implementations
