// Code generated by gormer. DO NOT EDIT.
package dao

import (
	"context"

	"gorm.io/gorm"
)

func getError(ctx context.Context, db *gorm.DB) (err error) {
	err = db.Error

	return
}
