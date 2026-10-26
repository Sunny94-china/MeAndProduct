package password

import "golang.org/x/crypto/bcrypt"

const DefaultBcryptCost = 14

// Hash 使用 bcrypt 生成密码哈希
func Hash(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), DefaultBcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 恒定时间比较明文和哈希
func Verify(plain, hash string) error {
	if plain == "" || hash == "" {
		return ErrEmptyPassword
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

var (
	ErrEmptyPassword = &passwordError{"password is empty"}
)

type passwordError struct{ msg string }

func (e *passwordError) Error() string { return e.msg }
