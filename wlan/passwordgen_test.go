package wlan

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func IsUnsupportedError(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "unsupported")
}

func Test_GeneratePassword_unknown_version(t *testing.T) {
	if _, err := GenerateDevicePassword(-1, 10); !IsUnsupportedError(err) {
		t.Error("Unknown versions should not be supported")
	}
}

func Test_GeneratePassword_invalid_length(t *testing.T) {
	if _, err := GenerateDevicePassword(1, -1); !IsUnsupportedError(err) {
		t.Error("Too small lengths should not be supported")
	}
	if _, err := GenerateDevicePassword(1, 0); !IsUnsupportedError(err) {
		t.Error("Too small lengths should not be supported")
	}
	if _, err := GenerateDevicePassword(1, 65); !IsUnsupportedError(err) {
		t.Error("Too large lengths should not be supported")
	}
}

func do_Test_GeneratePassword(version int, t *testing.T) {
	var err error
	_, err = os.Stat(efuseMacAddressFilename)
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
	_, err = os.Stat(usidFilename)
	if os.IsNotExist(err) {
		// Fallback for systems that don't have the original efuse file
		fp, err := ioutil.TempFile("", "dummy-usid")
		if err != nil {
			t.Fatalf("Could not create temporary file: %s", err)
		}
		usidFilename = fp.Name()
		defer os.Remove(usidFilename)

		_, err = fp.WriteString("dummy-usid\n")
		if err != nil {
			fp.Close()
			t.Fatalf("Could not write to temporary file: %s", err)
		}
		fp.Close()
	}

	// the same password is generated on multiple calls
	password1, err := GenerateDevicePassword(version, 10)
	if err != nil {
		t.Error("Could not generate password 1", err)
	}
	if len(password1) != 10 {
		t.Error("Password 1 has wrong length", password1)
	}
	password2, err := GenerateDevicePassword(version, 10)
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
	password3, err := GenerateDevicePassword(version, 5)
	if err != nil {
		t.Error("Could not generate password 3", err)
	}
	if password3 == "" {
		t.Error("Password 3 is empty")
	}
	if strings.HasPrefix(password1, password3) {
		t.Error("Password 3 should not be a prefix of password 1", password3, password1)
	}

	// check a single password with default length and code.
	password4Expected := "6cff1b65eb7d4f96"
	password4, err := GenerateDevicePassword(DefaultPasswordGeneratorVersion, DefaultPasswordLength)
	if err != nil {
		t.Error("Could not generate password 4", err)
	}
	if password4 != password4Expected {
		t.Error("Password 4 has not the expected walue", password4, password4Expected)
	}

}

func Test_GeneratePassword_v1(t *testing.T) {
	do_Test_GeneratePassword(1, t)
}
