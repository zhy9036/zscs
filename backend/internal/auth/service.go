package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zscaler/migration-platform/backend/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken     = errors.New("username already taken")
)

const (
	minPasswordLen = 8
	maxPasswordLen = 72
	minUsernameLen = 3
)

type Service struct {
	users      *user.Repository
	tokens     *TokenManager
	devRegister bool
}

func NewService(users *user.Repository, tokens *TokenManager, devRegister bool) *Service {
	return &Service{users: users, tokens: tokens, devRegister: devRegister}
}

type LoginResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	User        user.User
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	ok, err := VerifyPassword(password, u.PasswordHash)
	if err != nil || !ok {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, exp, err := s.tokens.Issue(u.ID, u.Username)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(exp).Seconds()),
		User:        u,
	}, nil
}

func (s *Service) Register(ctx context.Context, username, password string) (user.User, error) {
	if !s.devRegister {
		return user.User{}, errors.New("registration is disabled")
	}
	if err := validateCredentials(username, password); err != nil {
		return user.User{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return user.User{}, err
	}

	u, err := s.users.Create(ctx, username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return user.User{}, ErrUsernameTaken
		}
		return user.User{}, err
	}
	return u, nil
}

func validateCredentials(username, password string) error {
	if len(username) < minUsernameLen {
		return errors.New("username must be at least 3 characters")
	}
	if len(password) < minPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > maxPasswordLen {
		return errors.New("password is too long")
	}
	return nil
}
