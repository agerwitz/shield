package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/blowfish"
	"golang.org/x/crypto/twofish"
)

func Stream(enctype string, key, iv []byte) (cipher.Stream, cipher.Stream, error) {
	// cipher-mode combinations included so far are:
	// aes-cfb, blowfish-cfb, twofish-cfb
	if !strings.Contains(enctype, "-") {
		return nil, nil, errors.New("Invalid encryption type " + enctype + " specified")
	}
	cipherName := strings.Split(enctype, "-")[0]
	mode := strings.Split(enctype, "-")[1]

	var err error
	var block cipher.Block

	switch cipherName {
	// Was originally going to specify aes128 or aes256, but the keysize determines
	// which is used.
	case "aes128", "aes256":
		block, err = aes.NewCipher(key)
	case "blowfish":
		block, err = blowfish.NewCipher(key)
	case "twofish":
		block, err = twofish.NewCipher(key)
	default:
		return nil, nil, errors.New("Invalid cipher " + cipherName + " specified")
	}

	if err != nil {
		return nil, nil, err
	}

	switch mode {
	case "cfb":
		return cipher.NewCFBEncrypter(block, iv), cipher.NewCFBDecrypter(block, iv), nil
	case "ofb":
		return cipher.NewOFB(block, iv), cipher.NewOFB(block, iv), nil
	case "ctr":
		return cipher.NewCTR(block, iv), cipher.NewCTR(block, iv), nil
	default:
		return nil, nil, errors.New("Invalid encryption mode " + cipherName + " specified")
	}
}

func Initialize() (string, string, error) {
	//Generate the random IV
	iv := make([]byte, twofish.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", err
	}
	return "", "", nil
}
