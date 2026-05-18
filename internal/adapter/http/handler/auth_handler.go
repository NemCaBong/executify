package handler

import (
	"errors"
	"net/http"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/application/user"
	"github.com/NemCaBong/executify/pkg/httperr"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userUC *user.Usecase
}

func NewAuthHandler(userUC *user.Usecase) *AuthHandler {
	return &AuthHandler{userUC: userUC}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, err.Error())
		return
	}

	out, err := h.userUC.Register(c.Request.Context(), &user.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken), errors.Is(err, user.ErrUsernameTaken):
			httperr.Conflict(c, err.Error())
		default:
			httperr.Internal(c)
		}
		return
	}

	c.JSON(http.StatusCreated, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, err.Error())
		return
	}

	out, err := h.userUC.Login(c.Request.Context(), &user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCreds):
			httperr.Unauthorized(c, err.Error())
		case errors.Is(err, user.ErrInactiveUser):
			httperr.Forbidden(c, err.Error())
		default:
			httperr.Internal(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, err.Error())
		return
	}

	out, err := h.userUC.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidToken):
			httperr.Unauthorized(c, err.Error())
		case errors.Is(err, user.ErrInactiveUser):
			httperr.Forbidden(c, err.Error())
		default:
			httperr.Internal(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, err.Error())
		return
	}

	if err := h.userUC.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidToken):
			httperr.Unauthorized(c, err.Error())
		default:
			httperr.Internal(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
