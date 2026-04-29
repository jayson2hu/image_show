package service

import "errors"

var (
	ErrVerificationTooFrequent = errors.New("verification code sent too frequently")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrEmailExists             = errors.New("email already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrRegisterDisabled        = errors.New("registration is disabled")
	ErrUserDisabled            = errors.New("user is disabled")
)
