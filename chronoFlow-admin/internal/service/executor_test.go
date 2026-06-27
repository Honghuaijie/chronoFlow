package service

import (
	"context"
	"testing"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

func TestExecutorServiceCreateExecutor(t *testing.T) {
	repo := &serviceExecutorRepo{items: map[int64]*biz.Executor{}}
	uc := biz.NewExecutorUsecase(repo, serviceTokenCipher{}, log.DefaultLogger)
	svc := NewExecutorService(uc)

	reply, err := svc.CreateExecutor(context.Background(), &v1.CreateExecutorRequest{
		Name:        "exec-a",
		Address:     "http://127.0.0.1:19090",
		Token:       "secret",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("CreateExecutor returned error: %v", err)
	}
	if reply.GetCode() != 0 || reply.GetData().GetExecutor().GetId() != 1 {
		t.Fatalf("unexpected reply: %+v", reply)
	}
	if reply.GetData().GetExecutor().GetStatus() != biz.ExecutorStatusOffline {
		t.Fatalf("expected offline, got %q", reply.GetData().GetExecutor().GetStatus())
	}
}

type serviceExecutorRepo struct {
	items map[int64]*biz.Executor
}

func (r *serviceExecutorRepo) Create(_ context.Context, executor *biz.Executor) (*biz.Executor, error) {
	cp := *executor
	cp.ID = 1
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *serviceExecutorRepo) GetByID(_ context.Context, id int64) (*biz.Executor, error) {
	item := r.items[id]
	if item == nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (r *serviceExecutorRepo) GetByAddress(_ context.Context, address string) (*biz.Executor, error) {
	for _, item := range r.items {
		if item.Address == address {
			cp := *item
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *serviceExecutorRepo) List(_ context.Context) ([]*biz.Executor, error) {
	items := make([]*biz.Executor, 0, len(r.items))
	for _, item := range r.items {
		cp := *item
		items = append(items, &cp)
	}
	return items, nil
}

func (r *serviceExecutorRepo) Update(_ context.Context, executor *biz.Executor) (*biz.Executor, error) {
	cp := *executor
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *serviceExecutorRepo) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

type serviceTokenCipher struct{}

func (serviceTokenCipher) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (serviceTokenCipher) Decrypt(ciphertext string) (string, error) {
	return "dec:" + ciphertext, nil
}
