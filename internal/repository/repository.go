// Package repository Интерфейс репозитория и общие методы
package repository

import (
	"context"

	"github.com/s-turchinskiy/urlsavailability/models"
)

type Repository interface {
	Update(ctx context.Context, kit map[string]bool) (num uint64, err error)
	GetAllData(ctx context.Context) (models.FileStore, error)
	GetDataWithFilter(ctx context.Context, nums []uint64) (models.URLsKit, error)
	LoadAllData(ctx context.Context, data models.FileStore) error
}
