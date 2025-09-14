// internal/handlers/auth_handler.go
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
	"github.com/redis/go-redis/v9"
)

// ur : user repositories
type UserHandler struct {
	ur        *repositories.UserRepository
	authCache *utils.AuthCacheManager
}

func NewUserHandler(ur *repositories.UserRepository, rdb *redis.Client) *UserHandler {
	return &UserHandler{
		ur:        ur,
		authCache: utils.NewAuthCacheManager(rdb),
	}
}

// Register
// @Summary      Register a new user
// @Description  Create a user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      models.RegisterUser  true  "Register payload"
// @Success      201      {object}  map[string]interface{}  "Created user id and email"
// @Failure      400      {object}  map[string]string      "Bad request"
// @Failure      500      {object}  map[string]string      "Internal server error"
// @Router       /api/v1/register [post]
func (u *UserHandler) Register(ctx *gin.Context) {
	var user models.RegisterUser
	if err := ctx.ShouldBind(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Hash password
	// "password": "ceganssangar123(DF&&"
	// format : email + sangar123(DF&&
	hashCfg := pkg.NewHashConfig()
	hashCfg.UseRecommended()
	hashedPassword, err := hashCfg.GenHash(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// @Summary 	User Login
	newUser, err := u.ur.CreateUser(ctx, user.Email, hashedPassword)
	if err != nil {
		log.Println("error : ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to register"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":    newUser.Id,
		"email": newUser.Email,
	})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBind(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("email: ", user.Email)
	log.Println("password: ", user.Password)

	// GetID from Database
	infoUser, err := u.ur.GetIDFromEmail(ctx, user.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println("id : ", infoUser.Id)

	// Get password & role from where ID is match
	userCred, err := u.ur.GetPasswordFromID(ctx, infoUser.Id)
	if err != nil {
		log.Println("error getting password & role")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println("role : ", userCred.Role)
	log.Println("hashedPassword : ", userCred.Password)

	// Bandingkan password
	hashCfg := pkg.NewHashConfig()
	isMatched, err := hashCfg.CompareHashAndPassword(user.Password, userCred.Password)
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		re := regexp.MustCompile("hash|crypto|argon2id|format")
		if re.Match([]byte(err.Error())) {
			log.Println("Error during Hashing")
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	if !isMatched {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nama atau Password salah",
		})
		return
	}

	// Jika match, maka buatkan jwt dan kirim via response
	claims := pkg.NewJWTClaims(infoUser.Id, userCred.Role)
	jwtToken, err := claims.GenToken()
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   jwtToken,
	})
}

// --- Logout
func (u *UserHandler) Logout(ctx *gin.Context) {
	// Extract the token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		utils.HandleError(ctx, http.StatusBadRequest, "missing authorization header", "authorization header is required")
		return
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		utils.HandleError(ctx, http.StatusBadRequest, "invalid authorization format", "authorization header must be in format 'Bearer <token>'")
		return
	}

	// Get claims from context (set by middleware)
	claims, exists := ctx.Get("claims")
	if !exists {
		utils.HandleError(ctx, http.StatusUnauthorized, "unauthorized", "token claims not found")
		return
	}

	userClaims, ok := claims.(pkg.Claims)
	if !ok {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot cast claims")
		return
	}

	// Calculate remaining TTL for the token
	// Assuming your JWT has an expiration time
	// expirationTime := time.Unix(userClaims.ExpiresAt, 0)
	expirationTime := time.Unix(userClaims.ExpiresAt.Unix(), 0)
	remainingTTL := time.Until(expirationTime)

	// Only blacklist if token hasn't expired yet
	if remainingTTL > 0 {
		if err := u.authCache.BlacklistToken(ctx.Request.Context(), tokenString, remainingTTL); err != nil {
			utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to logout")
			return
		}
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data: map[string]interface{}{
			"message": "Logout successful",
		},
	})
}

// GetProfileByUserID handles GET /users/:user_id/profile
// It fetches the user's profile from the repository and returns it as JSON
func (u *UserHandler) GetProfile(ctx *gin.Context) {
	// Get the userID from token
	claims, _ := ctx.Get("claims")
	user, ok := claims.(pkg.Claims)
	if !ok {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot cast into pkg.claims")
		return
	}

	profile, err := u.ur.GetProfile(ctx, user.UserId)
	if err != nil {
		utils.HandleError(ctx, http.StatusOK, err.Error(), "profile not found")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    profile,
	})

}

func (u *UserHandler) EditProfile(ctx *gin.Context) {
	// Get image from form-data
	var body models.EditUserProfile
	log.Println("--- Begin ShouldBind")
	if err := ctx.ShouldBind(&body); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", err.Error())
		return
	}

	// Get the userID from token
	claims, _ := ctx.Get("claims")
	user, ok := claims.(pkg.Claims)
	if !ok {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot cast into pkg.claims")
		return
	}

	// Storing image process
	file := body.Img
	ext := filepath.Ext(file.Filename)
	re := regexp.MustCompile("(png|jpg|jpeg|webp)$")
	if !re.Match([]byte(ext)) {
		// Abort upload
	}
	filename := fmt.Sprintf("%d_images_%s%s", time.Now().UnixNano(), user.UserId, ext)
	location := filepath.Join("public", filename)
	if err := ctx.SaveUploadedFile(file, location); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to upload")
		return
	}

	log.Println("Success uploading image")
	log.Println("Location: ", location)

	editedProfile, err := u.ur.EditProfile(ctx.Request.Context(), user.UserId, body, location)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot edited user profile")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    editedProfile,
	})
}
