package model

import (
	"time"

	"github.com/glog"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	// for mysql
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql
)

var db *gorm.DB

// updateTimeStampForCreateCallback will set `CreatedAt`, `UpdatedAt` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Format("2006-01-02 15:04:05")
		if createTimeField, ok := scope.FieldByName("CreatedAt"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("UpdatedAt"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `UpdatedAt` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		// scope.SetColumn("UpdatedAt", time.Now().Unix())
		scope.SetColumn("UpdatedAt", time.Now().Format("2006-01-02 15:04:05"))
	}
}

// database migrate func
func migrate() {
	db.AutoMigrate(&HashRecord{})
}

// InitDataBase init mysql
func InitDataBase() {
	var (
		err                          error
		dbName, user, password, host string
	)
	dbName = viper.GetString("db.name")
	user = viper.GetString("db.username")
	password = viper.GetString("db.password")
	host = viper.GetString("db.addr")
	//password = "12345678"
	//host = "127.0.0.1:3306"

	//glog.Info("begin connect database...")
	//db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
	//	user,
	//	password,
	//	host,
	//	dbName))
	db, err = gorm.Open("mysql", user+":"+password+"@tcp("+host+")/"+dbName+"?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		glog.Errorf("open database err: %v", err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName
	}

	//println(tablePrefix)
	db.LogMode(false)
	db.SingularTable(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// CURD callback
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)

	migrate()

}
