package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

//create TABLE IF NOT EXISTS users(
//     chat_id INT unique,
//     token text unique
// );

type Postgres struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) NewUser(ctx context.Context, chatID int64, token string) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO users(chat_id, token) VALUES($1, $2)`

	_, err = tx.Exec(ctx, query, chatID, token)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) SaveToken(ctx context.Context, token string, chatId int64) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE users SET token = $1 WHERE chat_id = $2`

	_, err = tx.Exec(ctx, query, token)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) TakeToken(ctx context.Context, chatId int64) (string, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	query := `SELECT token FROM users WHERE chat_id = $1`

	var token string
	tx.QueryRow(ctx, query, chatId).Scan(&token)

	return token, nil
}
