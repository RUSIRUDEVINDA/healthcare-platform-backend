package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"healthcare-platform/services/auth-service/internal/model"
	"healthcare-platform/services/auth-service/internal/repository"
	"healthcare-platform/pkg/jwt"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
)

// Sentinel errors — handlers check these to return correct HTTP status codes
var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is deactivated. contact support")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthService contains all business logic for authentication
// It knows nothing about HTTP — that is the handler's job
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtHelper *jwt.Helper
	mqClient  *rabbitmq.Client
	log       *logger.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtHelper *jwt.Helper,
	mqClient *rabbitmq.Client,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtHelper: jwtHelper,
		mqClient:  mqClient,
		log:       log,
	}
}

// ──────────────────────────────────────────────
// Core Auth Operations
// ──────────────────────────────────────────────

func (s *AuthService) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	// Check if email is already taken
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("service.Register check email: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password with bcrypt cost=12 (good balance of security vs speed)
	// cost=12 takes ~250ms which makes brute-force attacks very slow
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("service.Register hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPwd),
		Role:         req.Role,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("service.Register save user: %w", err)
	}

	s.log.Info("New user registered", "user_id", user.ID, "role", user.Role, "email", user.Email)

	// Publish event to RabbitMQ
	// patient-service listens and creates a patient profile automatically
	// notification-service listens and sends a welcome email
	s.publishUserRegisteredEvent(user)

	return s.buildTokenResponse(user)
}

func (s *AuthService) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("service.Login find user: %w", err)
	}
	// Return same error for "not found" and "wrong password" 
	// This prevents email enumeration attacks
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Compare plaintext password against stored bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.log.Warn("Failed login attempt", "email", req.Email)
		return nil, ErrInvalidCredentials
	}

	s.log.Info("User logged in", "user_id", user.ID, "role", user.Role)

	return s.buildTokenResponse(user)
}

func (s *AuthService) RefreshToken(refreshTokenStr string) (*model.TokenResponse, error) {
	// Hash the incoming token to look it up in DB
	tokenHash := jwt.HashToken(refreshTokenStr)

	userID, err := s.userRepo.FindRefreshToken(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("service.RefreshToken find token: %w", err)
	}
	if userID == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Token rotation: delete old refresh token, issue new one
	// This limits the window of opportunity if a token is stolen
	if err := s.userRepo.DeleteRefreshToken(tokenHash); err != nil {
		s.log.Warn("Failed to delete old refresh token", "error", err)
	}

	s.log.Info("Token refreshed", "user_id", user.ID)
	return s.buildTokenResponse(user)
}

func (s *AuthService) Logout(refreshTokenStr string) error {
	tokenHash := jwt.HashToken(refreshTokenStr)
	if err := s.userRepo.DeleteRefreshToken(tokenHash); err != nil {
		return fmt.Errorf("service.Logout: %w", err)
	}
	return nil
}

// ValidateToken is called by the API Gateway to check if a JWT is valid
// Returns user info so downstream services know who the user is
func (s *AuthService) ValidateToken(tokenStr string) (*model.ValidateTokenResponse, error) {
	claims, err := s.jwtHelper.ParseAccessToken(tokenStr)
	if err != nil {
		return &model.ValidateTokenResponse{Valid: false}, nil
	}

	return &model.ValidateTokenResponse{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   model.Role(claims.Role),
	}, nil
}

// ──────────────────────────────────────────────
// Private helpers
// ──────────────────────────────────────────────

func (s *AuthService) buildTokenResponse(user *model.User) (*model.TokenResponse, error) {
	// Generate short-lived access token
	accessToken, expiresIn, err := s.jwtHelper.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("service.buildTokenResponse access token: %w", err)
	}

	// Generate long-lived refresh token (random UUID, not JWT)
	refreshToken := uuid.New().String() + uuid.New().String()
	tokenHash := jwt.HashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(time.Duration(s.jwtHelper.RefreshTTLDays) * 24 * time.Hour)

	if err := s.userRepo.SaveRefreshToken(user.ID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("service.buildTokenResponse save refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User: model.UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Role:       user.Role,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			IsVerified: user.IsVerified,
		},
	}, nil
}

func (s *AuthService) publishUserRegisteredEvent(user *model.User) {
	event := rabbitmq.UserRegisteredEvent{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	if err := s.mqClient.PublishUserRegistered(event); err != nil {
		// Non-fatal: log error but don't fail the registration
		// The user is registered, the event can be retried later
		s.log.Error("Failed to publish user.registered event", "user_id", user.ID, "error", err)
	} else {
		s.log.Info("Published user.registered event", "user_id", user.ID)
	}
}
