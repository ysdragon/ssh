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

// HashPasswordArgon2 hashes a password using Argon2id
func HashPasswordArgon2(password string, params *Argon2Params) (encodedHash string, err error) {
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

// ComparePasswordAndHash compares a password with an argon2 hash
func ComparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, nil // Not an argon2 hash
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

	expectedHash, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Generate the hash from the provided password
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))

	// Compare the hashes in constant time
	match = (subtle.ConstantTimeCompare(hash, expectedHash) == 1)

	return match, nil
}

// IsArgon2Hash detects if a string is an argon2 hash
func IsArgon2Hash(str string) bool {
	parts := strings.Split(str, "$")
	return len(parts) == 6 && parts[1] == "argon2id"
}
