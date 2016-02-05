package wlan

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
)

type devicePasswordGenerator interface {
	UpdateHash(h hash.Hash) error
}

const (
	passwordGenMagicHex = "c4f978b109c7a2c867ea856f677de245eb6fe63358a56314eebffb94009c51e9"
)

var (
	efuseMacAddressFilename = "/sys/class/efuse/mac"
	passwordGenMagic        []byte
	passwordGenerators      map[int]devicePasswordGenerator
)

// DefaultPasswordGeneratorVersion specifies the default generator version.
var DefaultPasswordGeneratorVersion = 1

// DefaultPasswordLength specifies the default length for password generator.
var DefaultPasswordLength = 16

func init() {
	var err error
	passwordGenMagic, err = hex.DecodeString(passwordGenMagicHex)
	if err != nil {
		panic(err)
	}

	passwordGenerators = map[int]devicePasswordGenerator{
		// Passwords of version "1" are generated from the efuse mac address only.
		1: &devicePasswordGeneratorMacOnly{},
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

type devicePasswordGeneratorMacOnly struct {
}

func (g *devicePasswordGeneratorMacOnly) UpdateHash(h hash.Hash) error {
	if err := addFileToHash(efuseMacAddressFilename, h); err != nil {
		return err
	}

	return nil
}

// GenerateDevicePassword generates a random device-specific password that
// can be used to protect a wifi network.
func GenerateDevicePassword(version int, length int) (string, error) {
	if length <= 0 || length > sha256.Size*2 {
		return "", errors.New("unsupported length")
	}

	generator, found := passwordGenerators[version]
	if !found {
		return "", errors.New("unsupported version")
	}

	h := hmac.New(sha256.New, passwordGenMagic)
	if err := addStringToHash(fmt.Sprintf("%d|%d", version, length), h); err != nil {
		return "", err
	}

	if err := generator.UpdateHash(h); err != nil {
		return "", err
	}

	hashsum := h.Sum(nil)
	return hex.EncodeToString(hashsum)[:length], nil
}
