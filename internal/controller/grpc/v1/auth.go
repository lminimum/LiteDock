package v1

import (
	"context"

	v1 "github.com/evrone/go-clean-template/docs/proto/v1"
	"github.com/evrone/go-clean-template/internal/entity"
)

// Login handles user login
func (r *V1) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	token, user, err := r.auth.Login(ctx, req.Username, req.Password)
	if err != nil {
		r.l.Error(err, "grpc - v1 - Login")
		return &v1.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.LoginResponse{
		Success: true,
		Token:   token,
		User:    convertEntityToProtoUser(user),
		Message: "Login successful",
	}, nil
}

// Register handles user registration
func (r *V1) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	role := req.Role
	if role == "" {
		role = "user"
	}

	user, err := r.auth.Register(ctx, req.Username, req.Password, req.Email, role)
	if err != nil {
		r.l.Error(err, "grpc - v1 - Register")
		return &v1.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		User:    convertEntityToProtoUser(user),
	}, nil
}

// GetCurrentUser retrieves the current authenticated user
func (r *V1) GetCurrentUser(ctx context.Context, req *v1.GetCurrentUserRequest) (*v1.GetCurrentUserResponse, error) {
	user, err := r.auth.GetCurrentUser(ctx, req.Token)
	if err != nil {
		r.l.Error(err, "grpc - v1 - GetCurrentUser")
		return &v1.GetCurrentUserResponse{
			Authenticated: false,
		}, nil
	}

	return &v1.GetCurrentUserResponse{
		Authenticated: true,
		User:          convertEntityToProtoUser(user),
	}, nil
}

// RefreshToken handles token refresh
func (r *V1) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, error) {
	newToken, err := r.auth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		r.l.Error(err, "grpc - v1 - RefreshToken")
		return &v1.RefreshTokenResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.RefreshTokenResponse{
		Success:     true,
		AccessToken: newToken,
		Message:     "Token refreshed successfully",
	}, nil
}

// convertEntityToProtoUser converts entity.User to proto.User
func convertEntityToProtoUser(user *entity.User) *v1.User {
	if user == nil {
		return nil
	}

	return &v1.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}
