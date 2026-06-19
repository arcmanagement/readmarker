package store

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var (
	bucketName        = []byte("cursors")
	ErrEmptySourceKey = errors.New("source_key must not be empty")
)

type Cursor struct {
	SourceKey string `json:"source_key"`
	Cursor    uint64 `json:"cursor"`
}

type Store struct {
	db *bbolt.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("db path must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	s := &Store{db: db}
	if err := s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize db: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Get(sourceKey string) (uint64, bool, error) {
	if sourceKey == "" {
		return 0, false, ErrEmptySourceKey
	}

	var cursor uint64
	var ok bool
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		value := bucket.Get([]byte(sourceKey))
		if value == nil {
			return nil
		}
		decoded, err := decodeCursor(value)
		if err != nil {
			return err
		}
		cursor = decoded
		ok = true
		return nil
	})
	return cursor, ok, err
}

func (s *Store) Advance(sourceKey string, next uint64) (uint64, error) {
	if sourceKey == "" {
		return 0, ErrEmptySourceKey
	}

	var final uint64
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		current, err := cursorFromBucket(bucket, sourceKey)
		if err != nil {
			return err
		}
		final = current
		if next > current {
			final = next
			return bucket.Put([]byte(sourceKey), encodeCursor(final))
		}
		return nil
	})
	return final, err
}

func (s *Store) Set(sourceKey string, cursor uint64) error {
	if sourceKey == "" {
		return ErrEmptySourceKey
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketName).Put([]byte(sourceKey), encodeCursor(cursor))
	})
}

func (s *Store) List() ([]Cursor, error) {
	var cursors []Cursor
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		return bucket.ForEach(func(key, value []byte) error {
			cursor, err := decodeCursor(value)
			if err != nil {
				return err
			}
			cursors = append(cursors, Cursor{
				SourceKey: string(key),
				Cursor:    cursor,
			})
			return nil
		})
	})
	return cursors, err
}

func cursorFromBucket(bucket *bbolt.Bucket, sourceKey string) (uint64, error) {
	value := bucket.Get([]byte(sourceKey))
	if value == nil {
		return 0, nil
	}
	return decodeCursor(value)
}

func encodeCursor(cursor uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, cursor)
	return buf
}

func decodeCursor(value []byte) (uint64, error) {
	if len(value) != 8 {
		return 0, fmt.Errorf("invalid cursor encoding: got %d bytes", len(value))
	}
	return binary.BigEndian.Uint64(value), nil
}
