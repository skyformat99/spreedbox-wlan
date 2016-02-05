package wlan

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io/ioutil"
)

const (
	passwordGenMagicHex = "c4f978b109c7a2c867ea856f677de245eb6fe63358a56314eebffb94009c51e9"
)

var (
	efuseMacAddressFilename = "/sys/class/efuse/mac"
	passwordGenMagic        []byte
)

func init() {
	var err error
	passwordGenMagic, err = hex.DecodeString(passwordGenMagicHex)
	if err != nil {
		panic(err)
	}
}

func addStringToHash(s string, h hash.Hash) error {
	h.Write([]byte(s))
	return nil
}

func addFileToHash(filename string, h hash.Hash) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("the file %s is empty", filename)
	}

	h.Write(data)
	return nil
}

// GenerateDevicePassword generates a random device-specific password that
// can be used to protect a wifi network.
func GenerateDevicePassword(length int) (string, error) {
	h := hmac.New(sha256.New, passwordGenMagic)
	if err := addStringToHash(fmt.Sprintf("%d", length), h); err != nil {
		return "", err
	}

	if err := addFileToHash(efuseMacAddressFilename, h); err != nil {
		return "", err
	}

	// TODO(jojo): add more files to hash

	hashsum := h.Sum(nil)
	return hex.EncodeToString(hashsum)[:length], nil
}
