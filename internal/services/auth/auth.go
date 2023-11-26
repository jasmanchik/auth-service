package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNoExists       = errors.New("user does not exists")
)

// New returns new instance of the Auth service
func New(
	log *slog.Logger,
	saver UserSaver,
	provider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    saver,
		userProvider: provider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, pass string, appID int32) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email), //careful with sensitive data!
		slog.Int("appID", int(appID)),
	)
	log.Info("start user login")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err.Error())

			return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", op, err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(pass))
	if err != nil {
		a.log.Error("invalid password", op, err)
		return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Error("app does not exists", op, err.Error())

		return "", fmt.Errorf("%s : %w", err)
	}
	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create jwt token", op, err.Error())

		return "", fmt.Errorf("%s : %w", err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email), //careful with sensitive data!
	)

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate a password hash", err.Error())
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	userID, err := a.userSaver.SaveUser(ctx, email, hash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user not found", err.Error())
			return 0, fmt.Errorf("%s : %w", op, ErrUserExists)
		}
		log.Error("failed to save user", err.Error())
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	return userID, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user id", userID), //careful with sensitive data!
	)
	log.Info("checking is user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("user not found", err.Error())
			return false, fmt.Errorf("%s : %w", op, ErrInvalidAppId)
		}

		return false, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("checked is user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
