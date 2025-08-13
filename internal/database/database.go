package database

import "context"

type Database interface {
	Save(ctx context.Context, tableName string, item interface{}) error
}
