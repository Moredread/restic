package crypto_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/restic/restic/internal/crypto"
	rtest "github.com/restic/restic/internal/test"
)

// TestZeroLengthPlaintext tests encryption and decryption of zero-length data
// This is an edge case that could cause issues in data handling
func TestZeroLengthPlaintext(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()

	plaintext := []byte{}
	ciphertext := k.Seal(nil, nonce, plaintext, nil)

	// Ciphertext should only contain MAC (16 bytes)
	rtest.Assert(t, len(ciphertext) == k.Overhead(),
		"ciphertext length incorrect for zero-length plaintext: want %d, got %d",
		k.Overhead(), len(ciphertext))

	// Decrypt and verify
	decrypted, err := k.Open(nil, nonce, ciphertext, nil)
	rtest.OK(t, err)
	rtest.Assert(t, len(decrypted) == 0,
		"decrypted zero-length plaintext should be empty, got %d bytes", len(decrypted))
}

// TestInvalidKeySeal tests that Seal panics with an invalid key
// This prevents silent corruption from using uninitialized keys
func TestInvalidKeySeal(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Seal with invalid key should panic")
		}
	}()

	// Create a zero key (invalid)
	k := &crypto.Key{}
	nonce := crypto.NewRandomNonce()
	data := []byte("test data")

	// This should panic
	_ = k.Seal(nil, nonce, data, nil)
}

// TestInvalidKeyOpen tests that Open returns error with an invalid key
// This prevents silent corruption from using uninitialized keys
func TestInvalidKeyOpen(t *testing.T) {
	// First create valid ciphertext
	validKey := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := []byte("test data")
	ciphertext := validKey.Seal(nil, nonce, data, nil)

	// Try to decrypt with invalid (zero) key
	invalidKey := &crypto.Key{}
	_, err := invalidKey.Open(nil, nonce, ciphertext, nil)

	rtest.Assert(t, err != nil, "Open with invalid key should return error")
}

// TestZeroNonceOpen tests that Open rejects zero nonces
// Zero nonces are a security vulnerability in CTR mode
func TestZeroNonceOpen(t *testing.T) {
	k := crypto.NewRandomKey()
	validNonce := crypto.NewRandomNonce()
	data := []byte("test data")
	ciphertext := k.Seal(nil, validNonce, data, nil)

	// Try to decrypt with zero nonce
	zeroNonce := make([]byte, k.NonceSize())
	_, err := k.Open(nil, zeroNonce, ciphertext, nil)

	rtest.Assert(t, err != nil, "Open with zero nonce should return error")
}

// TestCiphertextTooShort tests handling of truncated ciphertext
// This could indicate data corruption or incomplete writes
func TestCiphertextTooShort(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()

	testCases := []struct {
		name           string
		ciphertextLen  int
	}{
		{"empty", 0},
		{"one byte", 1},
		{"half MAC", k.Overhead() / 2},
		{"one byte short", k.Overhead() - 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create truncated ciphertext
			ciphertext := make([]byte, tc.ciphertextLen)

			_, err := k.Open(nil, nonce, ciphertext, nil)
			rtest.Assert(t, err != nil,
				"Open should fail with ciphertext of length %d", tc.ciphertextLen)
		})
	}
}

// TestCiphertextCorruptionAllBytes tests that corruption is detected in any byte
// This ensures MAC verification works correctly across the entire ciphertext
func TestCiphertextCorruptionAllBytes(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := rtest.Random(42, 256) // 256 bytes of data

	ciphertext := k.Seal(nil, nonce, data, nil)
	original := make([]byte, len(ciphertext))
	copy(original, ciphertext)

	// Test corruption of each byte in the ciphertext
	for i := 0; i < len(ciphertext); i++ {
		// Flip one bit
		ciphertext[i] ^= 0x01

		_, err := k.Open(nil, nonce, ciphertext, nil)
		if err != crypto.ErrUnauthenticated {
			t.Errorf("corrupted byte at position %d not detected: got error %v, want ErrUnauthenticated", i, err)
		}

		// Restore original byte
		ciphertext[i] = original[i]
	}
}

// TestMACCorruptionAllBytes tests that MAC corruption is detected in any MAC byte
// The MAC is critical for integrity - corruption of any bit should be detected
func TestMACCorruptionAllBytes(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := []byte("test data for MAC verification")

	ciphertext := k.Seal(nil, nonce, data, nil)
	original := make([]byte, len(ciphertext))
	copy(original, ciphertext)

	// MAC is the last k.Overhead() bytes
	macStart := len(ciphertext) - k.Overhead()

	// Test corruption of each byte in the MAC
	for i := macStart; i < len(ciphertext); i++ {
		for bit := uint(0); bit < 8; bit++ {
			// Flip one bit
			ciphertext[i] ^= (1 << bit)

			_, err := k.Open(nil, nonce, ciphertext, nil)
			if err != crypto.ErrUnauthenticated {
				t.Errorf("corrupted MAC bit at position %d, bit %d not detected: got error %v, want ErrUnauthenticated", i, bit, err)
			}

			// Restore original bit
			ciphertext[i] = original[i]
		}
	}
}

// TestNonceCorruption tests that using wrong nonce is detected
// Nonce mismatch should cause MAC verification failure
func TestNonceCorruption(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := []byte("test data")

	ciphertext := k.Seal(nil, nonce, data, nil)

	// Try decryption with different nonces
	for i := 0; i < 10; i++ {
		wrongNonce := crypto.NewRandomNonce()

		// Make sure we have a different nonce
		if bytes.Equal(wrongNonce, nonce) {
			continue
		}

		_, err := k.Open(nil, wrongNonce, ciphertext, nil)
		if err != crypto.ErrUnauthenticated {
			t.Errorf("wrong nonce not detected: got error %v, want ErrUnauthenticated", err)
		}
	}
}

// TestKeyCorruption tests that using wrong key is detected
// Key mismatch should cause MAC verification failure
func TestKeyCorruption(t *testing.T) {
	k1 := crypto.NewRandomKey()
	k2 := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := []byte("test data")

	// Encrypt with k1
	ciphertext := k1.Seal(nil, nonce, data, nil)

	// Try to decrypt with k2 (wrong key)
	_, err := k2.Open(nil, nonce, ciphertext, nil)
	if err != crypto.ErrUnauthenticated {
		t.Errorf("wrong key not detected: got error %v, want ErrUnauthenticated", err)
	}
}

// TestCiphertextTruncation tests detection of truncated ciphertext
// Truncation could occur during incomplete writes or storage errors
func TestCiphertextTruncation(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := rtest.Random(42, 1024)

	ciphertext := k.Seal(nil, nonce, data, nil)

	// Test truncation at various points
	truncationPoints := []int{
		len(ciphertext) - 1,     // one byte missing
		len(ciphertext) - 5,     // several bytes missing
		len(ciphertext) / 2,     // half the ciphertext
		k.Overhead() + 1,        // just past MAC
		k.Overhead(),            // exactly MAC size
	}

	for _, truncLen := range truncationPoints {
		t.Run(fmt.Sprintf("truncate_to_%d", truncLen), func(t *testing.T) {
			truncated := ciphertext[:truncLen]

			_, err := k.Open(nil, nonce, truncated, nil)
			rtest.Assert(t, err != nil,
				"truncated ciphertext (len=%d) should fail decryption", truncLen)
		})
	}
}

// TestCiphertextExtension tests handling of extended ciphertext
// Extra bytes could indicate data corruption or format errors
func TestCiphertextExtension(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()
	data := []byte("test data")

	ciphertext := k.Seal(nil, nonce, data, nil)

	// Append extra bytes
	extended := append(ciphertext, []byte("extra garbage data")...)

	// This should fail because MAC is calculated over the entire ciphertext
	// minus the MAC size, so extra bytes will be included in MAC verification
	_, err := k.Open(nil, nonce, extended, nil)
	if err != crypto.ErrUnauthenticated {
		t.Errorf("extended ciphertext not detected: got error %v, want ErrUnauthenticated", err)
	}
}

// TestPartiallyInvalidKey tests edge cases of key validity
func TestPartiallyInvalidKey(t *testing.T) {
	// Test encryption key valid, MAC key invalid
	k1 := &crypto.Key{}
	validKey := crypto.NewRandomKey()
	k1.EncryptionKey = validKey.EncryptionKey
	// MACKey is zero (invalid)

	rtest.Assert(t, !k1.Valid(), "key with zero MAC key should be invalid")

	// Test MAC key valid, encryption key invalid
	k2 := &crypto.Key{}
	k2.MACKey = validKey.MACKey
	// EncryptionKey is zero (invalid)

	rtest.Assert(t, !k2.Valid(), "key with zero encryption key should be invalid")
}

// TestNonceReuse tests the dangerous case of nonce reuse
// In CTR mode, nonce reuse can leak plaintext XOR
func TestNonceReuse(t *testing.T) {
	k := crypto.NewRandomKey()
	nonce := crypto.NewRandomNonce()

	data1 := []byte("first message with secret data")
	data2 := []byte("second message also has secrets")

	// Encrypt both with same nonce (this is a vulnerability)
	ciphertext1 := k.Seal(nil, nonce, data1, nil)
	ciphertext2 := k.Seal(nil, nonce, data2, nil)

	// Both should decrypt successfully (the vulnerability is subtle)
	decrypted1, err := k.Open(nil, nonce, ciphertext1, nil)
	rtest.OK(t, err)
	rtest.Equals(t, decrypted1, data1)

	decrypted2, err := k.Open(nil, nonce, ciphertext2, nil)
	rtest.OK(t, err)
	rtest.Equals(t, decrypted2, data2)

	// The vulnerability: XOR of ciphertexts reveals XOR of plaintexts
	// This test documents the risk of nonce reuse
	minLen := len(ciphertext1)
	if len(ciphertext2) < minLen {
		minLen = len(ciphertext2)
	}

	// Extract just the encrypted portion (without MAC)
	encData1 := ciphertext1[:len(ciphertext1)-k.Overhead()]
	encData2 := ciphertext2[:len(ciphertext2)-k.Overhead()]

	ctXor := make([]byte, len(encData1))
	for i := 0; i < len(encData1) && i < len(encData2); i++ {
		ctXor[i] = encData1[i] ^ encData2[i]
	}

	ptXor := make([]byte, len(data1))
	for i := 0; i < len(data1) && i < len(data2); i++ {
		ptXor[i] = data1[i] ^ data2[i]
	}

	// XOR of ciphertexts should equal XOR of plaintexts (the vulnerability)
	rtest.Assert(t, bytes.Equal(ctXor[:len(ptXor)], ptXor),
		"nonce reuse vulnerability: ciphertext XOR != plaintext XOR (this should match, showing the vulnerability)")
}
