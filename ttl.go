
/*
    установка TTL на каждый ключ
*/

package db

import (
	"errors"
	"time"
)

type ttl struct {
	Cache

	defaultTTL time.Duration
}

var ErrInvalidTTL = errors.New("TTL should be positive")

func newTtl(target Cache, defaultTtl time.Duration) (*ttl, error) {
	if defaultTtl < 0 {
		return nil, ErrInvalidTTL
	}
	ttl := &ttl{
		target,
		defaultTtl,
	}

	return ttl, nil
}

func (t *ttl) Get(key string) (*Value, error) {
	result, err := t.Cache.Get(key)
	if err != nil {
		return result, err
	}
	if result.Expires != 0 && result.Expires < time.Now().UnixNano() {
		return nil, ErrKeyNotFound
	}
	return result, err
}

func (t *ttl) Set(key string, value interface{}, expire time.Duration) (*Value, error) {

	var delay time.Duration = expire 

	if t.defaultTTL != 0 && expire == 0 {
		delay = t.defaultTTL
	}

	result, err := t.Cache.Set(key, value, delay)
	if err != nil {
		return nil, err
	}

	if delay > 0 {
		go t.delayRemove(key, result.Expires, delay)
	}
	return result, err
}

func (t *ttl) delayRemove(k string, controlExpire int64, delay time.Duration) {
	ticker := time.NewTicker(delay)
	<-ticker.C
	ticker.Stop()
	item, err := t.Cache.Get(k)
	if err != nil {
		return
	}
	if item.Expires == controlExpire {
		t.Cache.Remove(k)
	}
}
