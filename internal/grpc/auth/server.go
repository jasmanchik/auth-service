package auth

import (
	"context"
	"errors"
	ssov1 "github.com/jasmanchik/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/services/auth"
)

const emptyValue = 0

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int32,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Register(ctx context.Context, request *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {

	if request.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userId, err := s.auth.RegisterNewUser(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, request *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if request.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, request.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNoExists) {
			return nil, status.Error(codes.InvalidArgument, "invalid argument")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverAPI) Login(ctx context.Context, request *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {

	if request.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if request.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "app id is required")
	}

	token, err := s.auth.Login(ctx, request.GetEmail(), request.GetPassword(), request.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}
