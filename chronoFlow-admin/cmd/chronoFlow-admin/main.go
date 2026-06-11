package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/logger"
	"chronoFlow-admin/internal/worker"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

var (
	Name     = "chronoFlow-admin"
	Version  = "v0.1.0"
	flagconf string
	curEnv   string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path")
	flag.StringVar(&curEnv, "env", "local", "config env, eg: local")
}

func main() {
	flag.Parse()
	cfg, err := loadConfig(flagconf, curEnv)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	defer cfg.Close()

	var bc conf.Bootstrap
	if err := cfg.Scan(&bc); err != nil {
		log.Fatalf("failed to scan config: %v", err)
	}
	if err := conf.ValidateChronoFlow(&bc); err != nil {
		log.Fatalf("invalid ChronoFlow config: %v", err)
	}
	if err := setTimeZone(configuredTimezone(&bc)); err != nil {
		log.Fatalf("failed to set time zone: %v", err)
	}

	appLogger := logger.NewLogger(Name, bc.Logging)
	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Security, bc.Logs, bc.Executor, bc.Recovery, bc.Scheduler, appLogger)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer cleanup()
	if err := app.Run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}

func setTimeZone(tz string) error {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	time.Local = location
	return nil
}

func configuredTimezone(bc *conf.Bootstrap) string {
	if bc != nil && bc.Scheduler != nil && bc.Scheduler.Timezone != "" {
		return bc.Scheduler.Timezone
	}
	return "Asia/Shanghai"
}

func loadConfig(basePath, env string) (config.Config, error) {
	sources := []config.Source{
		file.NewSource(filepath.Join(basePath, "config.yaml")),
	}
	if env != "" {
		sources = append(sources, file.NewSource(filepath.Join(basePath, fmt.Sprintf("config-%s.yaml", env))))
	}

	c := config.New(config.WithSource(sources...))
	if err := c.Load(); err != nil {
		return nil, err
	}
	return c, nil
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, ws *worker.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs, ws),
	)
}
