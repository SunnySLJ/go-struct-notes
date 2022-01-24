package mysql

import (
	"fmt"
	"go-micro/internal/logx"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(user, password, ip string, port int, dbname string) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, ip, port, dbname)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// 结束后
	_ = DB.Callback().Create().After("gorm:after_create").Register(callBackLogName, afterLog)
	_ = DB.Callback().Query().After("gorm:after_query").Register(callBackLogName, afterLog)
	_ = DB.Callback().Delete().After("gorm:after_delete").Register(callBackLogName, afterLog)
	_ = DB.Callback().Update().After("gorm:after_update").Register(callBackLogName, afterLog)
	_ = DB.Callback().Row().After("gorm:row").Register(callBackLogName, afterLog)
	_ = DB.Callback().Raw().After("gorm:raw").Register(callBackLogName, afterLog)
	return
	return
}

const callBackLogName = "logx"

func afterLog(db *gorm.DB) {
	err := db.Error
	ctx := db.Statement.Context

	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	if err != nil {
		logx.WithTrace(ctx).Errorf("sql=%s || error=%v", sql, err)
		return
	}
	logx.WithTrace(ctx).Infof("sql=%s", sql)
}
