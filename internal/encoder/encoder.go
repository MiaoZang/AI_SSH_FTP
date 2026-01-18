package encoder

import (
	"encoding/base64"
	"fmt"
)

// Encode returns the base64 encoding of src.
func Encode(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}

// Decode decodes the base64 encoded src.
func Decode(src string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}
	return string(decoded), nil
}

// EncodeBytes returns the base64 encoding of src bytes.
func EncodeBytes(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// DecodeBytes decodes the base64 encoded src to bytes.
func DecodeBytes(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}
