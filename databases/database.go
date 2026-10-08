package databases

import (
	"fmt"
	"log"
	"time"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() error {
	cfg := configs.AppConfig

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FJakarta",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("⚠️ ERROR CONNECT TO DATABASE !", err)
	}

	// Konfigurasi connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("⚠️ Gagal mengambil *sql.DB dari GORM: ", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleTimeMin) * time.Minute)

	DB = db

	fmt.Printf("✅ CONNECT TO DATABASE COMPLETED! (pool: max_open=%d, max_idle=%d, lifetime=%dm, idle_time=%dm)\n",
		cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLifetimeMin, cfg.DBConnMaxIdleTimeMin)

	return nil
}