package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
)

// Sebelum review
// func VerifyToken(ctx *gin.Context) {
// 	// ambil token dari header
// 	bearerToken := ctx.GetHeader("Authorization")
// 	// Bearer token
// 	token := strings.Split(bearerToken, " ")[1]
// 	if token == "" {
// 		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 			"success": false,
// 			"error":   "Silahkan login terlebih dahulu",
// 		})
// 		return
// 	}

// 	var claims pkg.Claims
// 	if err := claims.VerifyToken(token); err != nil {
// 		if strings.Contains(err.Error(), jwt.ErrTokenInvalidIssuer.Error()) {
// 			log.Println("JWT Error.\nCause: ", err.Error())
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 				"success": false,
// 				"error":   "Silahkan login kembali",
// 			})
// 			return
// 		}
// 		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
// 			log.Println("JWT Error.\nCause: ", err.Error())
// 			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 				"success": false,
// 				"error":   "Silahkan login kembali",
// 			})
// 			return
// 		}
// 		fmt.Println(jwt.ErrTokenExpired)
// 		log.Println("Internal Server Error.\nCause: ", err.Error())
// 		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   "Internal Server Error",
// 		})
// 		return
// 	}
// 	ctx.Set("claims", claims)
// 	ctx.Next()
// }

func VerifyToken(ctx *gin.Context) {
	// ambil token dari header
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login terlebih dahulu", "Unauthorized Access")
		return
	}
	// Bearer token
	tokens := strings.Split(bearerToken, " ")
	if len(tokens) != 2 {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login terlebih dahulu", "Unauthorized Access")
		return
	}
	token := tokens[1]
	if token == "" {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login terlebih dahulu", "Unauthorized Access")
		return
	}

	var claims pkg.Claims
	if err := claims.VerifyToken(token); err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenInvalidIssuer.Error()) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "Invalid JWT")
			return
		}
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "Expired JWT")
			return
		}
		// fmt.Println(jwt.ErrTokenExpired)
		utils.HandleMiddlewareError(ctx, http.StatusInternalServerError, "Internal Server Error", "Internal Server Error")
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}
