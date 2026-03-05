package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
	"github.com/KaivalyaNaik/fitproof/pkg/email"
)

var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidOAuthState   = errors.New("invalid or expired oauth state")
	ErrInvalidVerifyCode   = errors.New("invalid or expired verification code")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

type RegisterResult struct {
	User         db.User
	AccessToken  string
	RefreshToken string
}

type LoginResult struct {
	User         db.User
	AccessToken  string
	RefreshToken string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type GoogleCallbackResult struct {
	User         db.User
	AccessToken  string
	RefreshToken string
	IsNewUser    bool
}

type AuthService struct {
	userRepo       *repositories.UserRepository
	tokenRepo      *repositories.TokenRepository
	emailVerifRepo *repositories.EmailVerificationRepository
	emailSender    *email.Sender
	oauthConfig    *oauth2.Config
	jwtSecret      []byte
	accessTTL      time.Duration
	refreshTTL     time.Duration
	emailVerifTTL  time.Duration
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	tokenRepo *repositories.TokenRepository,
	emailVerifRepo *repositories.EmailVerificationRepository,
	emailSender *email.Sender,
	oauthConfig *oauth2.Config,
	jwtSecret string,
	accessTTL, refreshTTL, emailVerifTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		emailVerifRepo: emailVerifRepo,
		emailSender:    emailSender,
		oauthConfig:    oauthConfig,
		jwtSecret:      []byte(jwtSecret),
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		emailVerifTTL:  emailVerifTTL,
	}
}

// ── token helpers ─────────────────────────────────────────────────────────────

func (s *AuthService) generateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateRawRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *AuthService) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func (s *AuthService) issueRefreshToken(ctx context.Context, userID uuid.UUID, deviceID, ip, ua string) (string, error) {
	raw, err := s.generateRawRefreshToken()
	if err != nil {
		return "", err
	}

	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if ua != "" {
		uaPtr = &ua
	}

	params := db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: s.hashToken(raw),
		DeviceID:  deviceID,
		IpAddress: ipPtr,
		UserAgent: uaPtr,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.refreshTTL), Valid: true},
	}

	_, err = s.tokenRepo.CreateRefreshToken(ctx, params)
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (s *AuthService) issueTokensForUser(ctx context.Context, user db.User, deviceID, ip, ua string) (string, string, error) {
	_ = s.tokenRepo.RevokeDeviceTokens(ctx, db.RevokeDeviceTokensParams{
		UserID:   user.ID,
		DeviceID: deviceID,
	})

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user.ID, deviceID, ip, ua)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ── password auth ─────────────────────────────────────────────────────────────

func (s *AuthService) Register(ctx context.Context, email, password, displayName, deviceID, ip, ua string) (RegisterResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return RegisterResult{}, err
	}

	hashStr := string(hash)
	user, err := s.userRepo.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: &hashStr,
		DisplayName:  displayName,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return RegisterResult{}, ErrEmailAlreadyExists
		}
		return RegisterResult{}, err
	}

	// Send email verification OTP — failure is non-fatal
	if sendErr := s.sendOTP(ctx, user); sendErr != nil {
		slog.Error("failed to send verification email", slog.String("error", sendErr.Error()))
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return RegisterResult{}, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user.ID, deviceID, ip, ua)
	if err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) Login(ctx context.Context, emailAddr, password, deviceID, ip, ua string) (LoginResult, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return LoginResult{}, ErrInvalidCredentials // OAuth-only account
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.issueTokensForUser(ctx, user, deviceID, ip, ua)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawToken, deviceID, ip, ua string) (RefreshResult, error) {
	hash := s.hashToken(rawToken)
	existing, err := s.tokenRepo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return RefreshResult{}, ErrInvalidRefreshToken
	}

	if existing.RevokedAt.Valid || existing.ExpiresAt.Time.Before(time.Now()) {
		return RefreshResult{}, ErrInvalidRefreshToken
	}

	if err := s.tokenRepo.RevokeRefreshToken(ctx, existing.ID); err != nil {
		return RefreshResult{}, err
	}

	accessToken, err := s.generateAccessToken(existing.UserID)
	if err != nil {
		return RefreshResult{}, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, existing.UserID, deviceID, ip, ua)
	if err != nil {
		return RefreshResult{}, err
	}

	return RefreshResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	hash := s.hashToken(rawToken)
	existing, err := s.tokenRepo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil
	}
	if existing.RevokedAt.Valid {
		return nil
	}
	return s.tokenRepo.RevokeRefreshToken(ctx, existing.ID)
}

// ── Google OAuth ──────────────────────────────────────────────────────────────

// GoogleLoginURL returns the Google consent URL. State is a short-lived signed JWT.
func (s *AuthService) GoogleLoginURL() (string, error) {
	nonce, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	now := time.Now()
	stateClaims := jwt.RegisteredClaims{
		Subject:   "oauth_state",
		ID:        nonce.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	}
	stateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, stateClaims)
	stateJWT, err := stateToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return s.oauthConfig.AuthCodeURL(stateJWT, oauth2.AccessTypeOnline), nil
}

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// HandleGoogleCallback exchanges the OAuth code, fetches the user profile,
// and finds or creates a user in the database.
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code, state, deviceID, ip, ua string) (GoogleCallbackResult, error) {
	// Validate state JWT
	parsed, err := jwt.ParseWithClaims(state, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return GoogleCallbackResult{}, ErrInvalidOAuthState
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject != "oauth_state" {
		return GoogleCallbackResult{}, ErrInvalidOAuthState
	}

	// Exchange code for token
	oauthToken, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return GoogleCallbackResult{}, fmt.Errorf("oauth exchange: %w", err)
	}

	// Fetch Google user info
	client := s.oauthConfig.Client(ctx, oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return GoogleCallbackResult{}, fmt.Errorf("userinfo fetch: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GoogleCallbackResult{}, fmt.Errorf("userinfo read: %w", err)
	}
	var gInfo googleUserInfo
	if err := json.Unmarshal(body, &gInfo); err != nil {
		return GoogleCallbackResult{}, fmt.Errorf("userinfo parse: %w", err)
	}

	var user db.User
	isNewUser := false

	// 1. Known Google ID → existing user
	user, err = s.userRepo.GetUserByGoogleID(ctx, gInfo.ID)
	if err == nil {
		// found — fall through to issue tokens
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return GoogleCallbackResult{}, err
	} else {
		// 2. Email exists → link google_id
		user, err = s.userRepo.GetUserByEmail(ctx, gInfo.Email)
		if err == nil {
			user, err = s.userRepo.UpdateUserGoogleID(ctx, user.ID, gInfo.ID)
			if err != nil {
				return GoogleCallbackResult{}, err
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			// 3. New user — create OAuth account (email auto-verified)
			displayName := gInfo.Name
			if displayName == "" {
				displayName = gInfo.Email
			}
			user, err = s.userRepo.CreateUserOAuth(ctx, db.CreateUserOAuthParams{
				Email:       gInfo.Email,
				DisplayName: displayName,
				GoogleID:    &gInfo.ID,
			})
			if err != nil {
				return GoogleCallbackResult{}, err
			}
			isNewUser = true
		} else {
			return GoogleCallbackResult{}, err
		}
	}

	accessToken, refreshToken, err := s.issueTokensForUser(ctx, user, deviceID, ip, ua)
	if err != nil {
		return GoogleCallbackResult{}, err
	}

	return GoogleCallbackResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
	}, nil
}

// ── email verification ────────────────────────────────────────────────────────

// sendOTP generates a 6-digit OTP, stores a bcrypt hash, and emails the raw code.
func (s *AuthService) sendOTP(ctx context.Context, user db.User) error {
	// Delete any outstanding tokens for this user
	_ = s.emailVerifRepo.DeleteAll(ctx, user.ID)

	// Generate 6-digit code
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return err
	}
	rawCode := fmt.Sprintf("%06d", n.Int64())

	codeHash, err := bcrypt.GenerateFromPassword([]byte(rawCode), 10)
	if err != nil {
		return err
	}

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(s.emailVerifTTL), Valid: true}
	if _, err = s.emailVerifRepo.Create(ctx, user.ID, string(codeHash), expiresAt); err != nil {
		return err
	}

	return s.emailSender.SendVerificationCode(ctx, user.Email, user.DisplayName, rawCode)
}

// SendVerificationCode (re)sends a verification OTP to the user's email.
func (s *AuthService) SendVerificationCode(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}
	return s.sendOTP(ctx, user)
}

// GetUserByID fetches a user — used by handlers after mutating state.
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

// VerifyEmail checks the OTP and marks the user's email as verified.
func (s *AuthService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	token, err := s.emailVerifRepo.GetLatest(ctx, userID)
	if err != nil {
		return ErrInvalidVerifyCode
	}
	if err := bcrypt.CompareHashAndPassword([]byte(token.Code), []byte(code)); err != nil {
		return ErrInvalidVerifyCode
	}
	if err := s.emailVerifRepo.MarkUsed(ctx, token.ID); err != nil {
		return err
	}
	return s.userRepo.SetEmailVerified(ctx, userID)
}
