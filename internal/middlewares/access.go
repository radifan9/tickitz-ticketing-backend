package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
)

// Sebelum review
// func Access(roles ...string) func(*gin.Context) {
// 	return func(ctx *gin.Context) {
// 		// Ambil data claim
// 		claims, isExist := ctx.Get("claims")
// 		if !isExist {
// 			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
// 				"success": false,
// 				"error":   "Silahkan login kembali",
// 			})
// 			return
// 		}
// 		user, ok := claims.(pkg.Claims)
// 		if !ok {
// 			// log.Println("Cannot cast claims into pkg.claims")
// 			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
// 				"success": false,
// 				"error":   "Internal server error",
// 			})
// 			return
// 		}
// 		if !slices.Contains(roles, user.Role) {
// 			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
// 				"success": false,
// 				"error":   "Anda tidak punya hak akses untuk resource ini",
// 			})
// 			return
// 		}
// 		ctx.Next()
// 	}
// }

func Access(roles ...string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// ambil data claim
		claims, isExist := ctx.Get("claims")
		if !isExist {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "silahkan login kembali", "Unauthorized Access")
			return
		}
		user, ok := claims.(pkg.Claims)
		if !ok {
			utils.HandleMiddlewareError(ctx, http.StatusInternalServerError, "Internal Server Error", "cannot cast into pkg.claims")
			return
		}
		if !slices.Contains(roles, user.Role) {
			utils.HandleMiddlewareError(ctx, http.StatusForbidden, "Anda tidak punya hak akses untuk resource ini", "Forbidden Access")
			return
		}
		ctx.Next()
	}
}
