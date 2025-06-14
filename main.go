package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"offair-cli/db"
	"offair-cli/menu"
)

func main() {
	// Load environment variables from .env file, ignoring any errors
	_ = godotenv.Load()

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create a bold cyan color function
	boldCyan := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Print welcome message with color
	fmt.Println(boldCyan("Welcome to OffAir CLI!"))
	menu.MainMenu(database)
}
