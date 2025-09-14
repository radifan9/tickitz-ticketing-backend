package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// AuthCacheManager handles authentication-related caching operations
type AuthCacheManager struct {
	rdb *redis.Client
}

// NewAuthCacheManager creates a new auth cache manager instance
func NewAuthCacheManager(rdb *redis.Client) *AuthCacheManager {
	return &AuthCacheManager{
		rdb: rdb,
	}
}

// BlacklistToken adds a token to the blacklist with a fixed TTL
func (a *AuthCacheManager) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	// Use a consistent key format for blacklisted tokens
	key := fmt.Sprintf("tickitz:blacklist:%s", tokenString)
	log.Println("tokenString : ", tokenString)

	// Store the token in blacklist with the remaining TTL
	// key : tickitz:blacklist:<tokenString>
	// value : blacklisted
	err := a.rdb.Set(ctx, key, "blacklisted", ttl).Err()
	if err != nil {
		log.Printf("Failed to blacklist token: %v", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	log.Printf("Token successfully blacklisted for %v", ttl)
	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (a *AuthCacheManager) IsTokenBlacklisted(ctx context.Context, tokenString string) bool {
	key := fmt.Sprintf("tickitz:blacklist:%s", tokenString)

	// Check if the key exists
	result := a.rdb.Exists(ctx, key)
	if result.Err() != nil {
		log.Printf("Error checking token blacklist: %v", result.Err())
		return false
	}

	exists := result.Val() > 0
	if exists {
		log.Printf("Token is blacklisted")
	}

	return exists
}

// BlacklistUserTokens blacklists all tokens for a specific user (useful for "logout from all devices")
// func (a *AuthCacheManager) BlacklistUserTokens(ctx context.Context, userID string, duration time.Duration) error {
// 	// This creates a user-level blacklist
// 	key := fmt.Sprintf("tickitz:user_blacklist:%s", userID)

// 	err := a.rdb.Set(ctx, key, time.Now().Unix(), duration).Err()
// 	if err != nil {
// 		log.Printf("Failed to blacklist user tokens: %v", err)
// 		return fmt.Errorf("failed to blacklist user tokens: %w", err)
// 	}

// 	log.Printf("All tokens for user %s blacklisted for %v", userID, duration)
// 	return nil
// }

// IsUserTokensBlacklisted checks if all tokens for a user should be considered invalid
func (a *AuthCacheManager) IsUserTokensBlacklisted(ctx context.Context, userID string, tokenIssuedAt time.Time) bool {
	key := fmt.Sprintf("tickitz:user_blacklist:%s", userID)

	result := a.rdb.Get(ctx, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			// Key doesn't exist, so user tokens are not blacklisted
			return false
		}
		log.Printf("Error checking user token blacklist: %v", result.Err())
		return false
	}

	// Get the blacklist timestamp
	blacklistTimestamp, err := result.Int64()
	if err != nil {
		log.Printf("Error parsing blacklist timestamp: %v", err)
		return false
	}

	// If the token was issued before the blacklist time, it should be considered invalid
	return tokenIssuedAt.Unix() < blacklistTimestamp
}

// ClearUserTokenBlacklist removes the user-level token blacklist (useful after password reset, etc.)
// func (a *AuthCacheManager) ClearUserTokenBlacklist(ctx context.Context, userID string) error {
// 	key := fmt.Sprintf("tickitz:user_blacklist:%s", userID)

// 	err := a.rdb.Del(ctx, key).Err()
// 	if err != nil {
// 		log.Printf("Failed to clear user token blacklist: %v", err)
// 		return fmt.Errorf("failed to clear user token blacklist: %w", err)
// 	}

// 	log.Printf("User token blacklist cleared for user %s", userID)
// 	return nil
// }
