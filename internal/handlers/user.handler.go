// internal/handlers/auth_handler.go
package handlers

import (
	"log"
	"net/http"
	"regexp"

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

// GetProfileByUserID handles GET /users/:user_id/profile
// It fetches the user's profile from the repository and returns it as JSON
func (u *UserHandler) GetProfileByUserID(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user_id is required",
		})
		return
	}

	profile, err := u.ur.GetProfileByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "profile not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}
