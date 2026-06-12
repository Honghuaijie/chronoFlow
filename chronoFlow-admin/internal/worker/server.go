package worker

import (
	"context"
	"time"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/logstore"
	"chronoFlow-admin/internal/scheduler"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

const unknownResultMessage = "执行器重启或失联，执行结果未知"

var ProviderSet = wire.NewSet(NewServer)

type Server struct {
	executorConf *conf.Executor
	recoveryConf *conf.Recovery
	logsConf     *conf.Logs
	executorRepo biz.ExecutorRepo
	logRepo      biz.JobLogMaintenanceRepo
	cipher       biz.TokenCipher
	healthClient biz.ExecutorHealthClient
	fileStore    *logstore.FileStore
	scheduler    *scheduler.Manager
	log          *log.Helper
	stop         chan struct{}
}

func NewServer(
	executorConf *conf.Executor,
	recoveryConf *conf.Recovery,
	logsConf *conf.Logs,
	executorRepo biz.ExecutorRepo,
	logRepo biz.JobLogMaintenanceRepo,
	cipher biz.TokenCipher,
	healthClient biz.ExecutorHealthClient,
	fileStore *logstore.FileStore,
	scheduler *scheduler.Manager,
	logger log.Logger,
) *Server {
	return &Server{
		executorConf: executorConf,
		recoveryConf: recoveryConf,
		logsConf:     logsConf,
		executorRepo: executorRepo,
		logRepo:      logRepo,
		cipher:       cipher,
		healthClient: healthClient,
		fileStore:    fileStore,
		scheduler:    scheduler,
		log:          log.NewHelper(logger),
		stop:         make(chan struct{}),
	}
}

func (s *Server) Start(ctx context.Context) error {
	if s.scheduler != nil {
		s.scheduler.Start()
	}
	go s.runStartupRecovery()
	go s.runHealthLoop()
	go s.runKillingTimeoutLoop()
	go s.runLogCleanupLoop()
	<-s.stop
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
	return nil
}

func (s *Server) runStartupRecovery() {
	delay := time.Duration(s.recoverySeconds()) * time.Second
	select {
	case <-time.After(delay):
		_ = s.logRepo.MarkAllActiveLogsFailed(context.Background(), unknownResultMessage)
	case <-s.stop:
	}
}

func (s *Server) runHealthLoop() {
	interval := time.Duration(s.healthIntervalSeconds()) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.checkExecutorsOnce(context.Background())
		case <-s.stop:
			return
		}
	}
}

func (s *Server) runKillingTimeoutLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.logRepo.MarkKillingTimeoutLogsFailed(context.Background(), s.killingTimeoutSeconds(), "终止超时")
		case <-s.stop:
			return
		}
	}
}

func (s *Server) runLogCleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredLogs(context.Background())
		case <-s.stop:
			return
		}
	}
}

func (s *Server) checkExecutorsOnce(ctx context.Context) error {
	executors, err := s.executorRepo.List(ctx)
	if err != nil {
		return err
	}
	for _, executor := range executors {
		token, err := s.cipher.Decrypt(executor.TokenCiphertext)
		if err != nil {
			return err
		}
		if err := s.healthClient.Health(ctx, executor.Address, token); err != nil {
			executor.HeartbeatFailCount++
			if executor.HeartbeatFailCount >= s.failThreshold() {
				executor.Status = biz.ExecutorStatusOffline
				_ = s.logRepo.MarkActiveLogsFailedByExecutorID(ctx, executor.ID, unknownResultMessage)
			}
			if _, err := s.executorRepo.Update(ctx, executor); err != nil {
				return err
			}
			continue
		}
		now := time.Now()
		executor.Status = biz.ExecutorStatusOnline
		executor.HeartbeatFailCount = 0
		executor.LastHeartbeatTime = &now
		if _, err := s.executorRepo.Update(ctx, executor); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) cleanupExpiredLogs(ctx context.Context) {
	if s.logsConf == nil || s.logsConf.RetentionDays <= 0 {
		return
	}
	paths, err := s.logRepo.DeleteExpiredLogs(ctx, s.logsConf.RetentionDays)
	if err != nil {
		s.log.Errorf("delete expired log metadata failed: %v", err)
		return
	}
	if s.fileStore == nil {
		return
	}
	for _, path := range paths {
		if err := s.fileStore.Delete(ctx, path); err != nil {
			s.log.Errorf("delete expired log file failed: path=%s err=%v", path, err)
		}
	}
}

func (s *Server) healthIntervalSeconds() int32 {
	if s.executorConf != nil && s.executorConf.HealthCheckIntervalSeconds > 0 {
		return s.executorConf.HealthCheckIntervalSeconds
	}
	return 10
}

func (s *Server) failThreshold() int32 {
	if s.executorConf != nil && s.executorConf.HealthCheckFailThreshold > 0 {
		return s.executorConf.HealthCheckFailThreshold
	}
	return 3
}

func (s *Server) recoverySeconds() int32 {
	if s.recoveryConf != nil && s.recoveryConf.StartupRunningGraceSeconds > 0 {
		return s.recoveryConf.StartupRunningGraceSeconds
	}
	return 120
}

func (s *Server) killingTimeoutSeconds() int32 {
	if s.recoveryConf != nil && s.recoveryConf.KillingTimeoutSeconds > 0 {
		return s.recoveryConf.KillingTimeoutSeconds
	}
	return 60
}
