package data

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"jieyuc.cn/muxu/studio/internal/conf"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet()

// Data holds the long-lived storage clients shared by repos.
type Data struct {
	db *gorm.DB
}

// NewData opens the database client and returns it with a cleanup function.
func NewData(c *conf.Data) (*Data, func(), error) {
	dc := c.GetDatabase()
	db, err := gorm.Open(postgres.Open(dc.GetSource()))
	if err != nil {
		return nil, nil, err
	}
	if dc.GetDebug() {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	// Auto migration is a convenience for local development. In production,
	// apply schema changes as a separate reviewed step instead.
	if dc.GetAutoMigrate() {
		if err := migrate(db); err != nil {
			defer func() {
				err := sqlDB.Close()
				if err != nil {
					log.Error("close database error, %s", err.Error())
				}
			}()

			return nil, nil, err
		}
	}
	cleanup := func() {
		log.Info("closing the data resources")
		if err := sqlDB.Close(); err != nil {
			log.Error("failed closing the database", "err", err)
		}
	}
	return &Data{db: db}, cleanup, nil
}

func migrate(db *gorm.DB) error {
	// 新增数据表后需要在这里注册数据表
	return db.AutoMigrate()
}
