package config

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params holds the parameters for argon2 hashing
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params provides default parameters for argon2 hashing
var DefaultArgon2Params = &Argon2Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPasswordArgon2id hashes a password using Argon2id
func HashPasswordArgon2id(password string, params *Argon2Params) (encodedHash string, err error) {
	// Generate a random salt
	salt := make([]byte, params.SaltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Encode the hash and salt as base64
	b64Salt := base64.StdEncoding.EncodeToString(salt)
	b64Hash := base64.StdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// HashPasswordArgon2i hashes a password using Argon2i
func HashPasswordArgon2i(password string, params *Argon2Params) (encodedHash string, err error) {
	// Generate a random salt
	salt := make([]byte, params.SaltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Generate the hash
	hash := argon2.Key([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Encode the hash and salt as base64
	b64Salt := base64.StdEncoding.EncodeToString(salt)
	b64Hash := base64.StdEncoding.EncodeToString(hash)

	// Format: $argon2i$v=19$m=65536,t=3,p=2$salt$hash
	encodedHash = fmt.Sprintf("$argon2i$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// addBase64Padding adds proper padding to a base64 string if needed
func addBase64Padding(s string) string {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return s
}

// ComparePasswordAndHash compares a password with an argon2 hash
func ComparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || (parts[1] != "argon2id" && parts[1] != "argon2i" && parts[1] != "argon2d") {
		// Return if the hash is not in the expected format
		return false, nil
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, nil
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	// Add proper padding to the base64 encoded hash before decoding
	paddedHash := addBase64Padding(parts[5])
	expectedHash, err := base64.StdEncoding.DecodeString(paddedHash)
	if err != nil {
		return false, err
	}

	// Generate the hash from the provided password based on the algorithm type
	var hash []byte
	switch parts[1] {
	case "argon2id":
		hash = argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))
	case "argon2i":
		hash = argon2.Key([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))
	case "argon2d":
		return false, fmt.Errorf("argon2d verification not supported - Go's argon2 package does not implement the Argon2d algorithm")
	default:
		return false, nil
	}

	// Compare the hashes in constant time
	match = (subtle.ConstantTimeCompare(hash, expectedHash) == 1)

	return match, nil
}

// IsArgon2Hash detects if a string is an argon2 hash
func IsArgon2Hash(str string) bool {
	parts := strings.Split(str, "$")
	return len(parts) == 6 && (parts[1] == "argon2id" || parts[1] == "argon2i" || parts[1] == "argon2d")
}
