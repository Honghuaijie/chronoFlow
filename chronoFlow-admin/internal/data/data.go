package data

import (
	"context"
	"fmt"
	"net/url"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewTransaction,
	NewUserRepo,
	NewExecutorRepo,
	NewJobRepo,
	NewGlueRepo,
	NewJobLogRepo,
	NewSystemSettingRepo,
	wire.Bind(new(biz.ExecutorRepo), new(*ExecutorRepo)),
	wire.Bind(new(biz.ExecutorLookupRepo), new(*ExecutorRepo)),
	wire.Bind(new(biz.JobRepo), new(*JobRepo)),
	wire.Bind(new(biz.JobLookupRepo), new(*JobRepo)),
	wire.Bind(new(biz.GlueRepo), new(*GlueRepo)),
	wire.Bind(new(biz.JobLogRepo), new(*JobLogRepo)),
	wire.Bind(new(biz.JobRunLogRepo), new(*JobLogRepo)),
	wire.Bind(new(biz.JobLogMaintenanceRepo), new(*JobLogRepo)),
	wire.Bind(new(biz.AlertJobLogRepo), new(*JobLogRepo)),
	wire.Bind(new(biz.SystemSettingRepo), new(*SystemSettingRepo)),
)

type Data struct {
	log *log.Helper
	db  *gorm.DB
}

func NewData(c *conf.Data, logger log.Logger, db *gorm.DB) (*Data, func(), error) {
	helper := log.NewHelper(logger)
	cleanup := func() {
		helper.Info("closing data resources")
	}
	return &Data{
		log: helper,
		db:  db,
	}, cleanup, nil
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	if c == nil || c.Database == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConfigInvalid, "数据库配置不能为空")
	}

	loc := url.QueryEscape(c.Database.Loc)
	parseTime := "False"
	if c.Database.ParseTime {
		parseTime = "True"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.Charset,
		parseTime,
		loc,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&User{}, &Executor{}, &Job{}, &JobGlue{}, &JobLog{}, &SystemSetting{}); err != nil {
		return nil, err
	}
	return db, nil
}

type contextTxKey struct{}

func NewTransaction(d *Data) biz.Transaction {
	return d
}

func (d *Data) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, contextTxKey{}, tx)
		return fn(txCtx)
	})
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.db
}
