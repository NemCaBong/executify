package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/NemCaBong/executify/internal/domain"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrInvalidToken  = errors.New("invalid or expired refresh token")
	ErrInactiveUser  = errors.New("user account is inactive")
)

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // access token TTL in seconds
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Usecase struct {
	userRepo         Repository
	refreshTokenRepo RefreshTokenRepository
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

func NewUsecase(
	userRepo Repository,
	refreshTokenRepo RefreshTokenRepository,
	jwtSecret []byte,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Usecase {
	return &Usecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

func (uc *Usecase) Register(ctx context.Context, in *RegisterInput) (*AuthOutput, error) {
	if existing, _ := uc.userRepo.GetByEmail(ctx, in.Email); existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.Create(ctx, &domain.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: string(hash),
		IsActive:     true,
	})
	if err != nil {
		return nil, err
	}

	return uc.issueTokenPair(ctx, user)
}

func (uc *Usecase) Login(ctx context.Context, in *LoginInput) (*AuthOutput, error) {
	user, err := uc.userRepo.GetByEmail(ctx, in.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCreds
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return uc.issueTokenPair(ctx, user)
}

// Refresh validates the incoming refresh token, revokes it, and issues a new token pair.
func (uc *Usecase) Refresh(ctx context.Context, rawToken string) (*AuthOutput, error) {
	hash := hashToken(rawToken)

	rt, err := uc.refreshTokenRepo.GetByTokenHash(ctx, hash)
	if err != nil || rt == nil || !rt.IsValid() {
		return nil, ErrInvalidToken
	}

	if err := uc.refreshTokenRepo.Revoke(ctx, rt.ID); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.GetByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	return uc.issueTokenPair(ctx, user)
}

// Logout revokes all refresh tokens for the user so every device is signed out.
func (uc *Usecase) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)

	rt, err := uc.refreshTokenRepo.GetByTokenHash(ctx, hash)
	if err != nil || rt == nil {
		return ErrInvalidToken
	}

	return uc.refreshTokenRepo.RevokeAllForUser(ctx, rt.UserID)
}

func (uc *Usecase) issueTokenPair(ctx context.Context, user *domain.User) (*AuthOutput, error) {
	accessToken, err := uc.signAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	if err := uc.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(uc.refreshTokenTTL),
	}); err != nil {
		return nil, err
	}

	return &AuthOutput{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(uc.accessTokenTTL.Seconds()),
	}, nil
}

func (uc *Usecase) signAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(uc.accessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(uc.jwtSecret)
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
