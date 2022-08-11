package postgres

import (
	"context"
	"database/sql"
	"time"
	
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vietquy/alpha/authn"
	"github.com/vietquy/alpha/errors"


)

var (
	errSave     = errors.New("failed to save key in database")
	errRetrieve = errors.New("failed to retrieve key from database")
	errDelete   = errors.New("failed to delete key from database")
)
var _ authn.KeyRepository = (*repo)(nil)

const (
	errDuplicate = "unique_violation"
	errInvalid   = "invalid_text_representation"
)

type repo struct {
	db *sqlx.DB
}

// New instantiates a PostgreSQL implementation of key repository.
func New(db *sqlx.DB) authn.KeyRepository {
	return &repo{
		db: db,
	}
}

func (kr repo) Save(ctx context.Context, key authn.Key) (string, error) {
	q := `INSERT INTO keys (id, type, issuer, issued_at, expires_at)
	      VALUES (:id, :type, :issuer, :issued_at, :expires_at)`

	dbKey := toDBKey(key)
	if _, err := kr.db.NamedExecContext(ctx, q, dbKey); err != nil {

		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code.Name() == errDuplicate {
				return "", errors.Wrap(authn.ErrConflict, pqErr)
			}
		}

		return "", errors.Wrap(errSave, err)
	}

	return dbKey.ID, nil
}

func (kr repo) Retrieve(ctx context.Context, issuer, id string) (authn.Key, error) {
	q := `SELECT id, type, issuer, issued_at, expires_at FROM keys WHERE issuer = $1 AND id = $2`
	key := dbKey{}
	if err := kr.db.QueryRowxContext(ctx, q, issuer, id).StructScan(&key); err != nil {
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return authn.Key{}, errors.Wrap(authn.ErrNotFound, err)
		}

		return authn.Key{}, errors.Wrap(errRetrieve, err)
	}

	return toKey(key), nil
}

func (kr repo) Remove(ctx context.Context, issuer, id string) error {
	q := `DELETE FROM keys WHERE issuer = :issuer AND id = :id`
	key := dbKey{
		ID:     id,
		Issuer: issuer,
	}
	if _, err := kr.db.NamedExecContext(ctx, q, key); err != nil {
		return errors.Wrap(errDelete, err)
	}

	return nil
}

type dbKey struct {
	ID        string       `db:"id"`
	Type      uint32       `db:"type"`
	Issuer    string       `db:"issuer"`
	Revoked   bool         `db:"revoked"`
	IssuedAt  time.Time    `db:"issued_at"`
	ExpiresAt sql.NullTime `db:"expires_at"`
}

func toDBKey(key authn.Key) dbKey {
	ret := dbKey{
		ID:       key.ID,
		Type:     key.Type,
		Issuer:   key.Issuer,
		IssuedAt: key.IssuedAt,
	}
	if !key.ExpiresAt.IsZero() {
		ret.ExpiresAt = sql.NullTime{Time: key.ExpiresAt, Valid: true}
	}

	return ret
}

func toKey(key dbKey) authn.Key {
	ret := authn.Key{
		ID:       key.ID,
		Type:     key.Type,
		Issuer:   key.Issuer,
		IssuedAt: key.IssuedAt,
	}
	if key.ExpiresAt.Valid {
		ret.ExpiresAt = key.ExpiresAt.Time
	}

	return ret
}
