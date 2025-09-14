package middlewares

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
	"github.com/redis/go-redis/v9"
)

var globalAuthCache *utils.AuthCacheManager

func InitAuthCache(rdb *redis.Client) {
	globalAuthCache = utils.NewAuthCacheManager(rdb)
}

// Verify Token without checking Redis Cache
func VerifyToken(ctx *gin.Context) {
	// Ambil token dari header
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

	// Check if token is blacklisted
	if globalAuthCache != nil {
		if globalAuthCache.IsTokenBlacklisted(ctx.Request.Context(), token) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "Token has been invalidated")
		}
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

	// Check if all user tokens are blacklisted (only if auth cache is initialized)
	if globalAuthCache != nil {
		if globalAuthCache.IsUserTokensBlacklisted(ctx.Request.Context(), claims.UserId, claims.IssuedAt.Time) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "All user tokens have been invalidated")
			return
		}
	}

	ctx.Set("claims", claims)
	ctx.Next()
}

// VerifyTokenWithBlacklist creates a middleware that checks both JWT validity and blacklist
func VerifyTokenWithBlacklist(rdb *redis.Client) gin.HandlerFunc {
	var authCache *utils.AuthCacheManager
	if rdb != nil {
		authCache = utils.NewAuthCacheManager(rdb)
		log.Println("VerifyTokenWithBlacklist: Auth cache initialized")
	} else {
		log.Println("VerifyTokenWithBlacklist: Warning - Redis client is nil")
	}

	return func(ctx *gin.Context) {
		// Ambil token dari header
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

		// Check if token is blacklisted (only if auth cache is available)
		if authCache != nil {
			if authCache.IsTokenBlacklisted(ctx.Request.Context(), token) {
				utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "Token has been invalidated")
				return
			}
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
			utils.HandleMiddlewareError(ctx, http.StatusInternalServerError, "Internal Server Error", "Internal Server Error")
			return
		}

		// Check if all user tokens are blacklisted (only if auth cache is available and claims are valid)
		if authCache != nil && claims.UserId != "" && claims.IssuedAt != nil && !claims.IssuedAt.IsZero() {
			if authCache.IsUserTokensBlacklisted(ctx.Request.Context(), claims.UserId, claims.IssuedAt.Time) {
				utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "All user tokens have been invalidated")
				return
			}
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
