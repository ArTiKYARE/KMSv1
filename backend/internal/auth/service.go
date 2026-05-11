package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kms-messenger/backend/internal/config"
	"github.com/kms-messenger/backend/internal/models"
	"github.com/kms-messenger/backend/internal/repository"
)

type AuthService struct {
	userRepo           *repository.UserRepository
	refreshTokenRepo   *repository.RefreshTokenRepository
	config             *config.Config
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*models.User, error) {
	existingUser, _ := s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	return s.userRepo.Create(ctx, email, password, firstName, lastName)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*Tokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.userRepo.VerifyPassword(user, password) {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	tokenRecord, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	return s.generateTokens(ctx, &tokenRecord.User)
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.DeleteByUserID(ctx, userID)
}

func (s *AuthService) generateTokens(ctx context.Context, user *models.User) (*Tokens, error) {
	accessToken, expiresAt, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	refreshTokenUUID := uuid.New().String()
	refreshExpiresAt := time.Now().Add(s.config.RefreshTokenExpiry)

	_, err = s.refreshTokenRepo.Create(ctx, user.ID, refreshTokenUUID, refreshExpiresAt)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenUUID,
		ExpiresIn:    int64(expiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

func (s *AuthService) generateJWT(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.config.JWTExpiry)

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
