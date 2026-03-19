package main

import (
	"LightObjS/internal/Auth"
	"LightObjS/internal/domain"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	migrationsPath     = "file://migrations"
	sqliteDbPath       = "./Data/users.db" // This should match dbPath in cmd/api/main.go
	superAdminUsername = "superadmin"
	superAdminPassword = "superadminpassword" // This should be set by user securely in production
)

func main() {
	fmt.Println("--- LightObjS Installation Utility ---")

	// Ensure the Data directory exists
	dataDir := "./Data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create Data directory: %v", err)
	}

	// 1. Run Migrations
	err := runMigrations(sqliteDbPath)
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	if err == migrate.ErrNoChange {
		fmt.Println("No database migrations to apply.")
	} else {
		fmt.Println("Database migrations applied successfully.")
	}

	// Initialize user repository for seeding
	userRepo, err := auth.NewSQLiteRepository(sqliteDbPath)
	if err != nil {
		log.Fatalf("Failed to initialize user repository for seeding: %v", err)
	}
	defer userRepo.Close()

	// 2. Seed Super Admin
	err = seedSuperAdmin(userRepo, superAdminUsername, superAdminPassword)
	if err != nil {
		log.Fatalf("Failed to seed super admin: %v", err)
	}
	fmt.Println("Super admin seeding complete.")

	fmt.Println("--- Installation Complete ---")
}

func runMigrations(dbPath string) error {
	m, err := migrate.New(
		migrationsPath,
		fmt.Sprintf("sqlite3://%s", dbPath),
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up() // Apply all available migrations
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return err
}

func seedSuperAdmin(userRepo auth.UserRepository, username, password string) error {
	// Check if super admin already exists
	existingUser, err := userRepo.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to check for existing super admin: %w", err)
	}
	if existingUser != nil {
		fmt.Printf("Super admin '%s' already exists, skipping seeding.\n", username)
		return nil
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash super admin password: %w", err)
	}

	// Create the super admin user
	superAdmin := &domain.User{
		Username:  username,
		Password:  hashedPassword, // Store hashed password
		CreatedAt: time.Now(),
		IsAdmin:   true, // Set super admin as an admin
	}

	if err := userRepo.CreateUser(superAdmin); err != nil {
		return fmt.Errorf("failed to create super admin user: %w", err)
	}

	fmt.Printf("Super admin '%s' seeded successfully.\n", username)
	return nil
}
