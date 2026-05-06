package handler

import (
	"errors"
	"net/http"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/application/user"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.userUC.Register(c.Request.Context(), &user.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, user.ErrEmailTaken) || errors.Is(err, user.ErrUsernameTaken) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.userUC.Login(c.Request.Context(), &user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, user.ErrInvalidCreds) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, user.ErrInactiveUser) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.userUC.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, user.ErrInvalidToken) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, user.ErrInactiveUser) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.NewAuthResponse(out.AccessToken, out.RefreshToken, out.ExpiresIn))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUC.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, user.ErrInvalidToken) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
