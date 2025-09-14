package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/radifan9/tickitz-ticketing-backend/internal/configs"
	"github.com/radifan9/tickitz-ticketing-backend/internal/routers"
)

// @title           Ticktiz Ticketing
// @version         1.0
// @description     RESTful API created using gin gonic
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	log.Println("--- --- Tickitz --- ---")

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("failed to load environment variables\nCause: ", err.Error())
		return
	}

	// PostgreSQL DB Initialization
	db, err := configs.InitDB()
	if err != nil {
		log.Println("failed to connect to database\nCause: ", err.Error())
		return
	}
	defer db.Close()

	// Test DB Connection
	if err := configs.TestDB(db); err != nil {
		log.Println("ping to DB failed\nCause: ", err.Error())
		return
	}
	log.Println("✅ PostgreSQL connected.")

	// Redis Initialization
	rdb := configs.InitRDB()
	defer rdb.Close()

	// Test Redis Connection
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Println("failed to ping redis database\nCause: ", err.Error())
		return
	}
	log.Println("✅ Successfully connect & ping to rdb!")

	// Engine Gin Initialization
	router := routers.InitRouter(db, rdb)
	router.Run(":3000")

	// Flow of the program
	// client => (router => handler => repo => handler) => client

}
