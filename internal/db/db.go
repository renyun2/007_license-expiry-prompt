package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"license-expiry/internal/models"
)

func Open(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_busy_timeout=5000", dbPath)
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.Certificate{},
		&models.ReminderConfig{},
		&models.RenewalApplication{},
		&models.AnnualInspectionRecord{},
		&models.FeeRecord{},
		&models.TodoItem{},
	)
}

func SeedFromFileIfEmpty(gdb *gorm.DB, initSQLPath string) error {
	var n int64
	if err := gdb.Model(&models.Certificate{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	sqlBytes, err := os.ReadFile(initSQLPath)
	if err != nil {
		log.Printf("seed skipped: no init.sql: %v", err)
		return nil
	}
	return gdb.Exec(string(sqlBytes)).Error
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
