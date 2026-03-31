package config

import (
	"log"
	"os"
	"path/filepath"

	"paygo/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	DBPath    string
	Port      string
	AdminUser string
	AdminPwd  string
	SysKey    string
}

var AppConfig *Config
var DB *gorm.DB

func LoadConfig(dbPath, port string) {
	AppConfig = &Config{
		DBPath:    dbPath,
		Port:      port,
		AdminUser: "admin",
		AdminPwd:  "12345678",
		SysKey:    "paygosyskey2024",
	}
}

func InitDB() {
	var err error
	dbPath := AppConfig.DBPath

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取底层sql.DB失败: %v", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 自动迁移
	err = DB.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.Record{},
		&model.Log{},
		&model.Order{},
		&model.RefundOrder{},
		&model.Settle{},
		&model.Batch{},
		&model.Transfer{},
		&model.PayType{},
		&model.Plugin{},
		&model.Channel{},
		&model.Roll{},
		&model.SubChannel{},
		&model.Config{},
		&model.Cache{},
		&model.Anounce{},
		&model.RegCode{},
		&model.InviteCode{},
		&model.Risk{},
		&model.Domain{},
		&model.Blacklist{},
		&model.PsReceiver{},
		&model.PsReceiver2{},
		&model.PsOrder{},
		&model.Weixin{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	log.Println("数据库初始化成功")
}
