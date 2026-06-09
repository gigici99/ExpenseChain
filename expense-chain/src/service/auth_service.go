package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// Claims is the JWT payload.
type Claims struct {
	UserID     string     `json:"user_id"`
	Username   string     `json:"username"`
	Role       model.Role `json:"role"`
	CompanyID  string     `json:"company_id,omitempty"`
	EmployeeID string     `json:"employee_id,omitempty"`
	TokenType  string     `json:"token_type"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// TokenPair is returned on login and refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type_header"` // always "Bearer"
}

type AuthService struct {
	repo   *repository.UserRepository
	secret []byte
}

func NewAuthService(repo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{repo: repo, secret: []byte(secret)}
}

// Register creates a new user with a bcrypt-hashed password.
func (s *AuthService) Register(username, password string, role model.Role, companyID, employeeID string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("AuthService.Register: username and password are required")
	}
	if role != model.RoleAdmin && role != model.RoleCompany && role != model.RoleEmployee {
		return nil, fmt.Errorf("AuthService.Register: invalid role %q", role)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: hash failed: %w", err)
	}

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CompanyID:    companyID,
		EmployeeID:   employeeID,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}
	return user, nil
}

// EnsureAdmin creates a seed ADMIN user at startup if it does not already exist.
// Idempotent: safe to call on every boot.
func (s *AuthService) EnsureAdmin(username, password string) error {
	if _, err := s.repo.FindByUsername(username); err == nil {
		log.Printf("[Auth] seed admin '%s' already exists", username)
		return nil
	}
	if _, err := s.Register(username, password, model.RoleAdmin, "", ""); err != nil {
		return fmt.Errorf("AuthService.EnsureAdmin: %w", err)
	}
	log.Printf("[Auth] seed admin '%s' created", username)
	return nil
}

// Login verifies credentials and issues an access+refresh token pair.
func (s *AuthService) Login(username, password string) (*TokenPair, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		// generic error — do not leak whether the username exists
		return nil, fmt.Errorf("AuthService.Login: invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("AuthService.Login: invalid credentials")
	}

	pair, err := s.issueTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Login: %w", err)
	}
	log.Printf("[Auth] login ok username=%s role=%s", user.Username, user.Role)
	return pair, nil
}

// Refresh validates a refresh token and issues a brand-new token pair.
func (s *AuthService) Refresh(refreshToken string) (*TokenPair, error) {
	claims, err := s.ParseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Refresh: %w", err)
	}
	if claims.TokenType != tokenTypeRefresh {
		return nil, fmt.Errorf("AuthService.Refresh: token is not a refresh token")
	}

	// reload user in case role/links changed
	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Refresh: %w", err)
	}

	pair, err := s.issueTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Refresh: %w", err)
	}
	log.Printf("[Auth] refreshed username=%s", user.Username)
	return pair, nil
}

// ParseToken validates signature + expiry and returns the claims.
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parse: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("token invalid")
	}
	return claims, nil
}

func (s *AuthService) issueTokenPair(user *model.User) (*TokenPair, error) {
	access, err := s.signToken(user, tokenTypeAccess, accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.signToken(user, tokenTypeRefresh, refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}, nil
}

func (s *AuthService) signToken(user *model.User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		CompanyID:  user.CompanyID,
		EmployeeID: user.EmployeeID,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
