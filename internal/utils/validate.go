package utils

import (
	"errors"
	"regexp"

	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

var (
	reMin8           = regexp.MustCompile(`^.{8,}$`)
	reMinSmall       = regexp.MustCompile(`[a-z]`)
	reMinLarge       = regexp.MustCompile(`[A-Z]`)
	reMinSpecialChar = regexp.MustCompile(`[!@#$%^&*/()]`)
)

func ValidatePassword(user models.ChangePasswordRequest) error {

	if !reMin8.MatchString(user.NewPassword) {
		return errors.New("password harus minimal 8 karakter")
	}
	if !reMinSmall.MatchString(user.NewPassword) {
		return errors.New("password minimal harus 1 karakter kecil")
	}
	if !reMinLarge.MatchString(user.NewPassword) {
		return errors.New("password minimal harus 1 karakter besar")
	}
	if !reMinSpecialChar.MatchString(user.NewPassword) {
		return errors.New("password harus ada karakter spesial")
	}
	return nil
}
