package repository

import "context"

type Repository interface {
	SaveToken(ctx context.Context, token string, chatId int64) error
	NewUser(ctx context.Context, chatID int64, token string) error
	TakeToken(ctx context.Context, chatId int64) (string, error)
}
