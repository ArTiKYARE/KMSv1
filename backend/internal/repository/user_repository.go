package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kms-messenger/backend/internal/database"
	"github.com/kms-messenger/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(ctx context.Context, email, password, firstName, lastName string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	result := database.GetDB().WithContext(ctx).Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := database.GetDB().WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := database.GetDB().WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (r *UserRepository) GetAll(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	result := database.GetDB().WithContext(ctx).
		Where("deleted_at IS NULL").
		Limit(limit).
		Offset(offset).
		Find(&users)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return users, nil
}

type RefreshTokenRepository struct{}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID string, token string, expiresAt time.Time) (*models.RefreshToken, error) {
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	result := database.GetDB().WithContext(ctx).Create(refreshToken)
	if result.Error != nil {
		return nil, result.Error
	}

	return refreshToken, nil
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	result := database.GetDB().WithContext(ctx).
		Preload("User").
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&refreshToken)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &refreshToken, nil
}

func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	result := database.GetDB().WithContext(ctx).Where("user_id = ?", userID).Delete(&models.RefreshToken{})
	return result.Error
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	result := database.GetDB().WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{})
	return result.Error
}
