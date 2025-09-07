// internal/handlers/auth_handler.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
)

// ur : user repositories
type UserHandler struct {
	ur *repositories.UserRepository
}

func NewUserHandler(ur *repositories.UserRepository) *UserHandler {
	return &UserHandler{ur: ur}
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var user models.RegisterUser
	if err := ctx.ShouldBind(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashCfg := pkg.NewHashConfig()
	hashCfg.UseRecommended()
	hashedPassword, err := hashCfg.GenHash(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Save user
	newUser, err := h.ur.CreateUser(ctx, user.Email, hashedPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":    newUser.Id,
		"email": newUser.Email,
	})
}
