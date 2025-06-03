package base

import (
	"gorm.io/gorm"
)

type TransactionalService struct {
	db *gorm.DB
}

func NewTransactionalService(db *gorm.DB) *TransactionalService {
	return &TransactionalService{db: db}
}

func (s *TransactionalService) WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
