package scheduler

import (
	"context"
	"sync"
	"time"

	"chronoFlow-admin/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

var ProviderSet = wire.NewSet(NewManager)

type RunFunc func(context.Context) error

type Manager struct {
	cron    *cron.Cron
	entries map[int64]cron.EntryID
	mu      sync.Mutex
	log     *log.Helper
}

func NewManager(c *conf.Scheduler) (*Manager, error) {
	locationName := "Asia/Shanghai"
	if c != nil && c.Timezone != "" {
		locationName = c.Timezone
	}
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, err
	}
	return &Manager{
		cron: cron.New(
			cron.WithLocation(location),
			cron.WithSeconds(),
			cron.WithChain(cron.Recover(cron.DefaultLogger)),
		),
		entries: make(map[int64]cron.EntryID),
		log:     log.NewHelper(log.DefaultLogger),
	}, nil
}

func (m *Manager) Start() {
	m.cron.Start()
}

func (m *Manager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
}

func (m *Manager) Register(jobID int64, cronExpr string, fn RunFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.entries[jobID]; ok {
		m.cron.Remove(existing)
		delete(m.entries, jobID)
	}
	entryID, err := m.cron.AddFunc(cronExpr, func() {
		if err := fn(context.Background()); err != nil {
			m.log.Errorf("cron job run failed: job_id=%d err=%v", jobID, err)
		}
	})
	if err != nil {
		return err
	}
	m.entries[jobID] = entryID
	return nil
}

func (m *Manager) Remove(jobID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entryID, ok := m.entries[jobID]
	if !ok {
		return
	}
	m.cron.Remove(entryID)
	delete(m.entries, jobID)
}

func (m *Manager) Has(jobID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.entries[jobID]
	return ok
}
