package db

import (
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"panic/internal/config"
	"panic/internal/domain"
)

var dbNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn, err := dataSource(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.DBDSN == "" {
		if err := createDatabase(cfg); err != nil {
			return nil, err
		}
	}

	level := logger.Warn
	if cfg.AppEnv == "development" {
		level = logger.Error
	}
	conn, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return conn, nil
}

func Migrate(conn *gorm.DB) error {
	return conn.AutoMigrate(
		&domain.User{},
		&domain.Seller{},
		&domain.Category{},
		&domain.Product{},
		&domain.Order{},
	)
}

func dataSource(cfg config.Config) (string, error) {
	if cfg.DBDSN != "" {
		return cfg.DBDSN, nil
	}
	mysqlCfg, err := mysqlConfig(cfg, true)
	if err != nil {
		return "", err
	}
	return mysqlCfg.FormatDSN(), nil
}

func createDatabase(cfg config.Config) error {
	if !dbNamePattern.MatchString(cfg.DBName) {
		return fmt.Errorf("DB_NAME hanya boleh huruf, angka, dan garis bawah")
	}
	mysqlCfg, err := mysqlConfig(cfg, false)
	if err != nil {
		return err
	}
	admin, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return err
	}
	defer admin.Close()
	_, err = admin.Exec("CREATE DATABASE IF NOT EXISTS `" + cfg.DBName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return err
}

func mysqlConfig(cfg config.Config, withDB bool) (mysqldriver.Config, error) {
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" {
		return mysqldriver.Config{}, fmt.Errorf("DB_HOST, DB_PORT, dan DB_USER wajib diisi")
	}
	c := mysqldriver.Config{
		User:                 cfg.DBUser,
		Passwd:               cfg.DBPassword,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Params:               map[string]string{"charset": "utf8mb4", "parseTime": "true", "loc": "Asia/Jakarta"},
		ParseTime:            true,
		AllowNativePasswords: true,
		Collation:            "utf8mb4_unicode_ci",
	}
	if withDB {
		if !dbNamePattern.MatchString(cfg.DBName) {
			return mysqldriver.Config{}, fmt.Errorf("DB_NAME hanya boleh huruf, angka, dan garis bawah")
		}
		c.DBName = cfg.DBName
	}
	return c, nil
}

func Ping(conn *gorm.DB) error {
	sqlDB, err := conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
