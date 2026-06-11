package biz

import (
	"context"
	"strings"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type Executor struct {
	ID                 int64
	Name               string
	Address            string
	TokenCiphertext    string
	Description        string
	Status             string
	HeartbeatFailCount int32
	LastHeartbeatTime  *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateExecutorInput struct {
	Name        string
	Address     string
	Token       string
	Description string
}

type UpdateExecutorInput struct {
	ID          int64
	Name        string
	Address     string
	Token       string
	Description string
}

type ExecutorRepo interface {
	Create(context.Context, *Executor) (*Executor, error)
	GetByID(context.Context, int64) (*Executor, error)
	List(context.Context) ([]*Executor, error)
	Update(context.Context, *Executor) (*Executor, error)
	Delete(context.Context, int64) error
}

type TokenCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type ExecutorUsecase struct {
	repo   ExecutorRepo
	cipher TokenCipher
	log    *log.Helper
}

func NewExecutorUsecase(repo ExecutorRepo, cipher TokenCipher, logger log.Logger) *ExecutorUsecase {
	return &ExecutorUsecase{
		repo:   repo,
		cipher: cipher,
		log:    log.NewHelper(logger),
	}
}

func (uc *ExecutorUsecase) CreateExecutor(ctx context.Context, input *CreateExecutorInput) (*Executor, error) {
	name := strings.TrimSpace(input.Name)
	address := strings.TrimRight(strings.TrimSpace(input.Address), "/")
	token := strings.TrimSpace(input.Token)
	if name == "" || address == "" || token == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "name、address 和 token 不能为空")
	}
	ciphertext, err := uc.cipher.Encrypt(token)
	if err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, &Executor{
		Name:            name,
		Address:         address,
		TokenCiphertext: ciphertext,
		Description:     strings.TrimSpace(input.Description),
		Status:          ExecutorStatusOffline,
	})
}

func (uc *ExecutorUsecase) UpdateExecutor(ctx context.Context, input *UpdateExecutorInput) (*Executor, error) {
	if input.ID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	existing, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "执行器不存在")
	}
	name := strings.TrimSpace(input.Name)
	address := strings.TrimRight(strings.TrimSpace(input.Address), "/")
	if name == "" || address == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "name 和 address 不能为空")
	}
	existing.Name = name
	existing.Address = address
	existing.Description = strings.TrimSpace(input.Description)
	if token := strings.TrimSpace(input.Token); token != "" {
		ciphertext, err := uc.cipher.Encrypt(token)
		if err != nil {
			return nil, err
		}
		existing.TokenCiphertext = ciphertext
	}
	return uc.repo.Update(ctx, existing)
}

func (uc *ExecutorUsecase) GetExecutor(ctx context.Context, id int64) (*Executor, error) {
	if id <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	executor, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if executor == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "执行器不存在")
	}
	return executor, nil
}

func (uc *ExecutorUsecase) ListExecutors(ctx context.Context) ([]*Executor, error) {
	return uc.repo.List(ctx)
}

func (uc *ExecutorUsecase) DeleteExecutor(ctx context.Context, id int64) error {
	if id <= 0 {
		return httpErrors.E(httpErrors.ErrInvalidID)
	}
	return uc.repo.Delete(ctx, id)
}
