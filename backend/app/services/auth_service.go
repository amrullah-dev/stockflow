package services

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractshttp "github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
)

var (
	ErrLoginEmailRequired      = errors.New("email is required")
	ErrLoginPasswordRequired   = errors.New("password is required")
	ErrLoginEmailInvalid       = errors.New("email must be valid")
	ErrLoginInvalidCredentials = errors.New("invalid credentials")
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRoleResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type LoginUserResponse struct {
	ID    uint64            `json:"id"`
	Name  string            `json:"name"`
	Email string            `json:"email"`
	Role  LoginRoleResponse `json:"role"`
}

type LoginResponse struct {
	Token string            `json:"token"`
	User  LoginUserResponse `json:"user"`
}

type AuthService struct{ db contractsorm.Orm }

func NewAuthService(db contractsorm.Orm) *AuthService { return &AuthService{db: db} }

func (s *AuthService) Login(ctx contractshttp.Context, input LoginInput) (*LoginResponse, error) {
	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" {
		return nil, ErrLoginEmailRequired
	}
	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email {
		return nil, ErrLoginEmailInvalid
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, ErrLoginPasswordRequired
	}

	var response *LoginResponse
	err = s.db.Transaction(func(tx contractsorm.Query) error {
		user, err := repositories.NewUserRepository(tx).FindByEmail(input.Email)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return ErrLoginInvalidCredentials
			}
			return err
		}
		if !user.IsActive || !facades.Hash().Check(input.Password, user.Password) {
			return ErrLoginInvalidCredentials
		}
		if err := repositories.NewUserRepository(tx).UpdateLastLoginAt(user, time.Now()); err != nil {
			return err
		}
		token, err := facades.Auth(ctx).Guard("user").LoginUsingID(user.ID)
		if err != nil {
			return err
		}
		response = &LoginResponse{Token: token, User: loginUserResponse(*user)}
		return nil
	})
	return response, err
}

func loginUserResponse(user models.User) LoginUserResponse {
	return LoginUserResponse{ID: uint64(user.ID), Name: user.Name, Email: user.Email, Role: LoginRoleResponse{ID: uint64(user.Role.ID), Name: user.Role.Name}}
}
