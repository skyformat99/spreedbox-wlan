package wlan

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_GeneratePassword(t *testing.T) {
	_, err := os.Stat(efuseMacAddressFilename)
	if os.IsNotExist(err) {
		// Fallback for systems that don't have the original efuse file
		fp, err := ioutil.TempFile("", "dummy-mac")
		if err != nil {
			t.Fatalf("Could not create temporary file: %s", err)
		}
		efuseMacAddressFilename = fp.Name()
		defer os.Remove(efuseMacAddressFilename)

		_, err = fp.WriteString("dummy-mac-address\n")
		if err != nil {
			fp.Close()
			t.Fatalf("Could not write to temporary file: %s", err)
		}
		fp.Close()
	}

	// the same password is generated on multiple calls
	password1, err := GenerateDevicePassword(10)
	if err != nil {
		t.Error("Could not generate password 1", err)
	}
	if len(password1) != 10 {
		t.Error("Password 1 has wrong length", password1)
	}
	password2, err := GenerateDevicePassword(10)
	if err != nil {
		t.Error("Could not generate password 2", err)
	}
	if len(password2) != 10 {
		t.Error("Password 2 has wrong length", password2)
	}
	if password1 != password2 {
		t.Error("passwords should be equal", password1, password2)
	}

	// the length is part of the generated password
	password3, err := GenerateDevicePassword(5)
	if err != nil {
		t.Error("Could not generate password 3", err)
	}
	if password3 == "" {
		t.Error("Password 3 is empty")
	}
	if strings.HasPrefix(password1, password3) {
		t.Error("Password 3 should not be a prefix of password 1", password3, password1)
	}
}
