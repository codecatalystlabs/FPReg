package service

import (
	"errors"
	"time"

	"fpreg/internal/config"
	"fpreg/internal/models"
	"fpreg/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.RefreshTokenRepository
	cfg       *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.RefreshTokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type AccessClaims struct {
	UserID     uuid.UUID   `json:"user_id"`
	Email      string      `json:"email"`
	Role       models.Role `json:"role"`
	FacilityID *uuid.UUID  `json:"facility_id,omitempty"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(email, password, ip, ua string) (*TokenPair, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}
	if !user.IsActive {
		return nil, nil, errors.New("account is deactivated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	pair, err := s.generateTokenPair(user, ip, ua)
	if err != nil {
		return nil, nil, err
	}
	return pair, user, nil
}

func (s *AuthService) RefreshTokens(refreshToken, ip, ua string) (*TokenPair, error) {
	rt, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil || rt.Revoked || rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invalid or expired refresh token")
	}

	if err := s.tokenRepo.Revoke(rt.ID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil || !user.IsActive {
		return nil, errors.New("user not found or inactive")
	}

	return s.generateTokenPair(user, ip, ua)
}

func (s *AuthService) Logout(refreshToken string) error {
	rt, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil {
		return nil
	}
	return s.tokenRepo.Revoke(rt.ID)
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService) generateTokenPair(user *models.User, ip, ua string) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(s.cfg.JWTAccessExpiryMinutes) * time.Minute)

	claims := AccessClaims{
		UserID:     user.ID,
		Email:      user.Email,
		Role:       user.Role,
		FacilityID: user.FacilityID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "fpreg",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshStr := uuid.New().String()
	refreshExp := now.Add(time.Duration(s.cfg.JWTRefreshExpiryHours) * time.Hour)

	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshStr,
		ExpiresAt: refreshExp,
		IPAddress: ip,
		UserAgent: ua,
	}
	if err := s.tokenRepo.Create(&rt); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessExp.Unix(),
	}, nil
}
