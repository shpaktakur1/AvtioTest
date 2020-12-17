
/*
Внешний интерфейс - Cache
*/

package db

import (
	"errors"
	"io"
	"time"
)

type Cache interface {
	Set(key string, value interface{}, expire time.Duration) (*Value, error)
	Get(key string) (*Value, error)
	Remove(key string) (err error)
	Keys() ([]string, error)
	GetAtIndex(key string, index interface{}) (interface{}, error)
}

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidValueType = errors.New("store only supports strings, lists and maps as values")
	ErrIndexAccess      = errors.New("cant Get item at index")
	ErrConversionError  = errors.New("failed to convert item")
	ErrIllegalIndexType = errors.New("list does not support given index type")
	ErrNonIntegerSubkey = errors.New("index conversion to int failed")
)

type DataType int

const (
	STRING DataType = iota
	LIST
	MAP
)

type Value struct {
	Type    DataType    `json:"type"`
	Data    interface{} `json:"data"`
	Expires int64       `json:"expires,omitempty"`
}

func NewCache(defaultTTL time.Duration, out io.Writer, rw io.ReadWriter, saveFreq time.Duration, nShards int, shardingFunc shardFunction) (c Cache, err error) {
	if nShards < 1 {
		nShards = 1
	}
	c, err = newSharder(nShards, shardingFunc)
	if err != nil {
		return nil, err
	}
	if out != nil {
		c = newLogger(c, out)
	}
	if rw != nil {
		c, err = newPersister(c, rw, saveFreq)
		if err != nil {
			return nil, err
		}
	}
	c, err = newTtl(c, defaultTTL)
	if err != nil {
		return nil, err
	}
	return
}
