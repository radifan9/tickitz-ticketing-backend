package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	// Daftar origin yang diizinkan
	whitelist := []string{
		"http://127.0.0.1:5500",
		"http://localhost:5173",
	}

	origin := ctx.GetHeader("Origin")
	if slices.Contains(whitelist, origin) {
		ctx.Header("Access-Control-Allow-Origin", origin)
	}

	// Header CORS standar
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	ctx.Header("Access-Control-Allow-Credentials", "true")

	// Jika request adalah preflight (OPTIONS)
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	// Lanjutkan ke handler berikutnya
	ctx.Next()
}
