package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHasher_DefaultCost(t *testing.T) {
	hasher := NewPasswordHasher(0)

	assert.NotNil(t, hasher)
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}

func TestNewPasswordHasher_ValidCost(t *testing.T) {
	cost := 12
	hasher := NewPasswordHasher(cost)

	assert.NotNil(t, hasher)
	assert.Equal(t, cost, hasher.cost)
}

func TestNewPasswordHasher_TooLowCost(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost - 1)

	assert.NotNil(t, hasher)
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}

func TestNewPasswordHasher_TooHighCost(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MaxCost + 1)

	assert.NotNil(t, hasher)
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}

func TestPasswordHasher_Hash(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost) // Use MinCost for faster tests

	password := "mySecurePassword123"
	hash, err := hasher.Hash(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash should be different from original
}

func TestPasswordHasher_Hash_DifferentHashes(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	password := "samePassword"
	hash1, _ := hasher.Hash(password)
	hash2, _ := hasher.Hash(password)

	// Bcrypt should generate different hashes for the same password
	// (due to random salt)
	assert.NotEqual(t, hash1, hash2)
}

func TestPasswordHasher_Compare_Success(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	password := "myPassword123"
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Compare(hash, password)
	assert.NoError(t, err)
}

func TestPasswordHasher_Compare_WrongPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	password := "correctPassword"
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Compare(hash, "wrongPassword")
	assert.Error(t, err)
}

func TestPasswordHasher_Compare_EmptyPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	password := "somePassword"
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Compare(hash, "")
	assert.Error(t, err)
}

func TestPasswordHasher_Compare_InvalidHash(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	err := hasher.Compare("not-a-valid-hash", "password")
	assert.Error(t, err)
}

func TestPasswordHasher_Hash_EmptyPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	// Bcrypt can hash empty strings
	hash, err := hasher.Hash("")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = hasher.Compare(hash, "")
	assert.NoError(t, err)
}

func TestPasswordHasher_Hash_LongPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	// Bcrypt has a max password length of 72 bytes
	// Passwords longer than 72 bytes will cause an error in newer bcrypt versions
	longPassword := "a"
	for i := 0; i < 100; i++ {
		longPassword += "a"
	}

	hash, err := hasher.Hash(longPassword)
	// bcrypt returns error for passwords > 72 bytes
	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestPasswordHasher_Hash_UnicodePassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	password := "пароль密码パスワード"
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Compare(hash, password)
	assert.NoError(t, err)
}

func TestPasswordHasher_Hash_SpecialCharacters(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)

	passwords := []string{
		"pass!@#$%^&*()",
		"pass with spaces",
		"pass\twith\ttabs",
		"pass\nwith\nnewlines",
		`pass"with'quotes`,
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := hasher.Hash(password)
			assert.NoError(t, err)

			err = hasher.Compare(hash, password)
			assert.NoError(t, err)
		})
	}
}

// Benchmark tests
func BenchmarkPasswordHasher_Hash_MinCost(b *testing.B) {
	hasher := NewPasswordHasher(bcrypt.MinCost)
	password := "benchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(password)
	}
}

func BenchmarkPasswordHasher_Hash_DefaultCost(b *testing.B) {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)
	password := "benchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(password)
	}
}

func BenchmarkPasswordHasher_Compare(b *testing.B) {
	hasher := NewPasswordHasher(bcrypt.MinCost)
	password := "benchmarkPassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Compare(hash, password)
	}
}

