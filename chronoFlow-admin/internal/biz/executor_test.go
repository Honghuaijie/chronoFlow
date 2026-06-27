package biz

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeExecutorRepo struct {
	created *Executor
	updated *Executor
	items   map[int64]*Executor
}

func (r *fakeExecutorRepo) Create(_ context.Context, executor *Executor) (*Executor, error) {
	cp := *executor
	cp.ID = 1
	r.created = &cp
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeExecutorRepo) GetByID(_ context.Context, id int64) (*Executor, error) {
	item := r.items[id]
	if item == nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (r *fakeExecutorRepo) GetByAddress(_ context.Context, address string) (*Executor, error) {
	for _, item := range r.items {
		if item.Address == address {
			cp := *item
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeExecutorRepo) List(_ context.Context) ([]*Executor, error) {
	items := make([]*Executor, 0, len(r.items))
	for _, item := range r.items {
		cp := *item
		items = append(items, &cp)
	}
	return items, nil
}

func (r *fakeExecutorRepo) Update(_ context.Context, executor *Executor) (*Executor, error) {
	cp := *executor
	r.updated = &cp
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeExecutorRepo) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

type fakeTokenCipher struct{}

func (fakeTokenCipher) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (fakeTokenCipher) Decrypt(ciphertext string) (string, error) {
	return "dec:" + ciphertext, nil
}

func TestExecutorUsecaseCreateEncryptsTokenAndDefaultsOffline(t *testing.T) {
	repo := &fakeExecutorRepo{items: map[int64]*Executor{}}
	uc := NewExecutorUsecase(repo, fakeTokenCipher{}, log.DefaultLogger)

	got, err := uc.CreateExecutor(context.Background(), &CreateExecutorInput{
		Name:        "  exec-a  ",
		Address:     " http://127.0.0.1:19090/ ",
		Token:       "secret-token",
		Description: "  desc  ",
	})
	if err != nil {
		t.Fatalf("CreateExecutor returned error: %v", err)
	}
	if got.ID != 1 {
		t.Fatalf("expected id=1, got %d", got.ID)
	}
	if repo.created.Name != "exec-a" || repo.created.Address != "http://127.0.0.1:19090" {
		t.Fatalf("executor fields were not normalized: %+v", repo.created)
	}
	if repo.created.TokenCiphertext != "enc:secret-token" {
		t.Fatalf("expected encrypted token, got %q", repo.created.TokenCiphertext)
	}
	if repo.created.Status != ExecutorStatusOffline {
		t.Fatalf("expected default offline, got %q", repo.created.Status)
	}
}

func TestExecutorUsecaseCreateRejectsDuplicateAddress(t *testing.T) {
	repo := &fakeExecutorRepo{items: map[int64]*Executor{
		7: {ID: 7, Name: "exists", Address: "http://127.0.0.1:19090"},
	}}
	uc := NewExecutorUsecase(repo, fakeTokenCipher{}, log.DefaultLogger)

	_, err := uc.CreateExecutor(context.Background(), &CreateExecutorInput{
		Name:    "new",
		Address: " http://127.0.0.1:19090/ ",
		Token:   "secret-token",
	})
	if err == nil {
		t.Fatal("expected duplicate address error, got nil")
	}
}

func TestExecutorUsecaseUpdateKeepsTokenWhenEmpty(t *testing.T) {
	now := time.Now()
	repo := &fakeExecutorRepo{items: map[int64]*Executor{
		1: {
			ID:                1,
			Name:              "old",
			Address:           "http://old",
			TokenCiphertext:   "enc:old",
			Description:       "old desc",
			Status:            ExecutorStatusOnline,
			LastHeartbeatTime: &now,
		},
	}}
	uc := NewExecutorUsecase(repo, fakeTokenCipher{}, log.DefaultLogger)

	got, err := uc.UpdateExecutor(context.Background(), &UpdateExecutorInput{
		ID:          1,
		Name:        "new",
		Address:     "http://new",
		Description: "new desc",
	})
	if err != nil {
		t.Fatalf("UpdateExecutor returned error: %v", err)
	}
	if got.TokenCiphertext != "enc:old" {
		t.Fatalf("expected old token ciphertext to be kept, got %q", got.TokenCiphertext)
	}
	if repo.updated.Status != ExecutorStatusOnline {
		t.Fatalf("expected status to be kept, got %q", repo.updated.Status)
	}
}

func TestExecutorUsecaseUpdateRejectsDuplicateAddress(t *testing.T) {
	now := time.Now()
	repo := &fakeExecutorRepo{items: map[int64]*Executor{
		1: {
			ID:                1,
			Name:              "old",
			Address:           "http://old",
			TokenCiphertext:   "enc:old",
			Status:            ExecutorStatusOnline,
			LastHeartbeatTime: &now,
		},
		2: {ID: 2, Name: "other", Address: "http://other"},
	}}
	uc := NewExecutorUsecase(repo, fakeTokenCipher{}, log.DefaultLogger)

	_, err := uc.UpdateExecutor(context.Background(), &UpdateExecutorInput{
		ID:      1,
		Name:    "new",
		Address: "http://other/",
	})
	if err == nil {
		t.Fatal("expected duplicate address error, got nil")
	}
}
