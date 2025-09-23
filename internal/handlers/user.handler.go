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
	ur *repositories.UserRepository
	ac *utils.AuthCacheManager
}

func NewUserHandler(ur *repositories.UserRepository, rdb *redis.Client) *UserHandler {
	return &UserHandler{
		ur: ur,
		ac: utils.NewAuthCacheManager(rdb),
	}
}

// @Summary Register a new user
// @Tags    Auth
// @Accept  json
// @Produce json
// @Param   body body models.RegisterUser true "User registration"
// @Success 201 {object} models.User
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/auth/register [post]
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
		utils.HandleError(ctx, http.StatusInternalServerError, "failed to hash password", err.Error())
		return
	}

	newUser, err := u.ur.CreateUser(ctx, user.Email, hashedPassword)
	if err != nil {
		log.Println("error : ", err)
		utils.HandleError(ctx, http.StatusConflict, "failed to register", err.Error())
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data: gin.H{
			"id":    newUser.Id,
			"email": newUser.Email,
		},
	})
}

// @Summary User login
// @Tags    Auth
// @Accept  json
// @Produce json
// @Param   body body models.User true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/auth/login [post]
func (u *UserHandler) Login(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBind(&user); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", err.Error())
		return
	}

	log.Println("email: ", user.Email)
	log.Println("password: ", user.Password)

	// GetID from Database
	infoUser, err := u.ur.GetIDFromEmail(ctx, user.Email)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", err.Error())
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

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data: models.SuccessLoginResponse{
			Role:  userCred.Role,
			Token: jwtToken,
		},
	})
}

// @Summary User logout
// @Tags    Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/auth/logout [delete]
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
	expirationTime := time.Unix(userClaims.ExpiresAt.Unix(), 0)
	remainingTTL := time.Until(expirationTime)

	log.Println("expirationTime : ", expirationTime)
	log.Println("remainingTTL : ", remainingTTL)

	// Only blacklist if token hasn't expired yet
	if remainingTTL > 0 {
		if err := u.ac.BlacklistToken(ctx.Request.Context(), tokenString, remainingTTL); err != nil {
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

// @Summary Get user profile
// @Tags    Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserProfile
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/users/profile [get]
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

// @Summary Edit user profile
// @Tags    Users
// @Accept  multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param   first_name formData string false "First name"
// @Param   last_name formData string false "Last name"
// @Param   phone_number formData string false "Phone number"
// @Param   img formData file false "Profile image"
// @Success 200 {object} models.UserProfile
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/users/profile [patch]
func (u *UserHandler) EditProfile(ctx *gin.Context) {
	// Get image from form-data
	var body models.EditUserProfile
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

	// Dari postman harus ambil gambar baru
	file := body.Img
	if file != nil {
		ext := filepath.Ext(file.Filename)
		re := regexp.MustCompile(`(?i)\.(png|jpg|jpeg|webp)$`)
		if !re.MatchString(ext) {
			utils.HandleError(ctx, http.StatusBadRequest, "invalid file type", "only png, jpg, jpeg, webp allowed")
			return
		}

		filename := fmt.Sprintf("%d_images_%s%s", time.Now().UnixNano(), user.UserId, ext)
		location := filepath.Join("public/profile_pics", filename)

		if err := ctx.SaveUploadedFile(file, location); err != nil {
			utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to upload")
			return
		}

		// Update profile with new image
		editedProfile, err := u.ur.EditProfile(ctx.Request.Context(), user.UserId, body, filename)
		if err != nil {
			utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "cannot edit user profile")
			return
		}

		utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{Success: true, Status: http.StatusOK, Data: editedProfile})
		return
	}

	// If no image uploaded, just update profile without image
	editedProfile, err := u.ur.EditProfile(ctx.Request.Context(), user.UserId, body, "")
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot edit user profile")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{Success: true, Status: http.StatusOK, Data: editedProfile})

}

// @Summary Change user password
// @Tags    Users
// @Accept  json
// @Produce json
// @Security BearerAuth
// @Param   body body models.ChangePasswordRequest true "Password change"
// @Success 200 {object} map[string]string "Password changed"
// @Router  ${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1/users/password [patch]
func (u *UserHandler) ChangePassword(ctx *gin.Context) {
	var req models.ChangePasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", err.Error())
		return
	}

	// Get user ID from token claims
	claims, _ := ctx.Get("claims")
	userClaims, ok := claims.(pkg.Claims)
	if !ok {
		utils.HandleError(ctx, http.StatusUnauthorized, "unauthorized", "invalid claims")
		return
	}

	// Validate Password
	// If there's no error in data-binding, validate the password
	if err := utils.ValidatePassword(req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", err.Error())
		return
	}

	// Fetch current password hash from DB
	userCred, err := u.ur.GetPasswordFromID(ctx, userClaims.UserId)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", "failed to fetch user credentials")
		return
	}

	// Compare old password
	hashCfg := pkg.NewHashConfig()
	isMatched, err := hashCfg.CompareHashAndPassword(req.OldPassword, userCred.Password)
	if err != nil || !isMatched {
		utils.HandleError(ctx, http.StatusUnauthorized, "unauthorized", "old password does not match")
		return
	}

	// Hash new password
	hashCfg.UseRecommended()
	hashedPassword, err := hashCfg.GenHash(req.NewPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to hash new password")
		return
	}

	// Update password
	if err := u.ur.UpdatePassword(ctx, userClaims.UserId, hashedPassword); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", err.Error())
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data: map[string]string{
			"message": "Password changed successfully",
		},
	})
}
