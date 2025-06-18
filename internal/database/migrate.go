package database

import (
	"fmt"
	"log"

	"github.com/nguyendkn/nginx-manager/internal/models"
	"gorm.io/gorm"
)

// AllModels returns all models for migration
func AllModels() []interface{} {
	return []interface{}{
		&models.User{},
		&models.Certificate{},
		&models.AccessList{},
		&models.AccessListItem{},
		&models.ProxyHost{},
		&models.RedirectionHost{},
		&models.Stream{},
		&models.DeadHost{},
		&models.AuditLog{},
		&models.Token{},
		&models.Setting{},
		&models.NginxConfig{},
		&models.ConfigVersion{},
		&models.ConfigBackup{},
		&models.ConfigTemplate{},
		&models.ConfigApproval{},
	}
}

// AutoMigrate runs auto migration for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database auto-migration...")

	for _, model := range AllModels() {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	log.Println("Database auto-migration completed successfully")
	return nil
}

// SeedData creates initial data in the database
func SeedData(db *gorm.DB) error {
	log.Println("Seeding initial data...")

	// Create default admin user if not exists
	if err := createDefaultAdmin(db); err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	// Create default settings
	if err := createDefaultSettings(db); err != nil {
		return fmt.Errorf("failed to create default settings: %w", err)
	}

	log.Println("Data seeding completed successfully")
	return nil
}

// createDefaultAdmin creates a default admin user
func createDefaultAdmin(db *gorm.DB) error {
	// Check if any user exists
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	// If users exist, skip creating default admin
	if count > 0 {
		log.Println("Users already exist, skipping default admin creation")
		return nil
	}

	// Create default admin user
	admin := &models.User{
		Email:    "admin@example.com",
		Name:     "Administrator",
		Nickname: "Admin",
		Password: "changeme", // This will be hashed in BeforeCreate hook
		Roles:    models.StringArray{string(models.RoleAdmin)},
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}

	log.Printf("Default admin user created with email: %s", admin.Email)
	return nil
}

// createDefaultSettings creates default system settings
func createDefaultSettings(db *gorm.DB) error {
	defaultSettings := []models.Setting{
		{
			ID:   "default-site",
			Name: "Default Site",
			Value: models.JSON{
				"value": "Congratulations! You have successfully installed Nginx Proxy Manager.",
			},
		},
		{
			ID:   "disable-ipv6",
			Name: "Disable IPv6",
			Value: models.JSON{
				"value": false,
			},
		},
		{
			ID:   "cloudflare-api-token",
			Name: "Cloudflare API Token",
			Value: models.JSON{
				"value": "",
			},
		},
		{
			ID:   "default-intermediate-cert",
			Name: "Default Intermediate Certificate",
			Value: models.JSON{
				"value": "",
			},
		},
	}

	for _, setting := range defaultSettings {
		// Check if setting already exists
		var existing models.Setting
		err := db.Where("id = ?", setting.ID).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// If setting doesn't exist, create it
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&setting).Error; err != nil {
				return err
			}
			log.Printf("Created default setting: %s", setting.Name)
		}
	}

	return nil
}

// DropAllTables drops all tables (use with caution!)
func DropAllTables(db *gorm.DB) error {
	log.Println("WARNING: Dropping all tables...")

	// Drop tables in reverse order to avoid foreign key constraints
	models := AllModels()
	for i := len(models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(models[i]); err != nil {
			return fmt.Errorf("failed to drop table for %T: %w", models[i], err)
		}
	}

	log.Println("All tables dropped successfully")
	return nil
}

// ResetDatabase drops all tables and recreates them with seed data
func ResetDatabase(db *gorm.DB) error {
	log.Println("Resetting database...")

	// Drop all tables
	if err := DropAllTables(db); err != nil {
		return err
	}

	// Run auto-migration
	if err := AutoMigrate(db); err != nil {
		return err
	}

	// Seed initial data
	if err := SeedData(db); err != nil {
		return err
	}

	log.Println("Database reset completed successfully")
	return nil
}

// CheckDatabaseHealth performs basic health checks on the database
func CheckDatabaseHealth(db *gorm.DB) error {
	// Check if we can perform a simple query
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if all tables exist
	for _, model := range AllModels() {
		if !db.Migrator().HasTable(model) {
			return fmt.Errorf("table for model %T does not exist", model)
		}
	}

	return nil
}
