package db

import (
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

var ErrLessThanOneShard = errors.New("must have at least one shard")

type shardFunction func(string) uint32

type sharder struct {
	fn shardFunction

	needLock bool
	shards   []Cache
	locks    []sync.RWMutex
}


func Shard(function shardFunction, needLock bool, caches ...Cache) (s Cache, err error) {
	n := len(caches)
	if n < 1 {
		err = ErrLessThanOneShard
		return
	}
	if function == nil {
		function = defaultHash
	}

	s = &sharder{
		shards:   caches,
		needLock: needLock,
		locks:    make([]sync.RWMutex, n),
		fn:       function,
	}
	return
}

func newSharder(n int, function shardFunction) (s *sharder, err error) {
	if n < 1 {
		err = ErrLessThanOneShard
		return
	}

	if function == nil {
		function = defaultHash
	}

	stores := make([]Cache, n)
	for i := 0; i < n; i++ {
		stores[i] = newStore()
	}
	wrapped, err := Shard(function, true, stores...)
	return wrapped.(*sharder), err
}

func (s *sharder) getTargetShardIdx(key string) uint32 {
	if len(s.shards) == 1 {
		return 0
	}
	return s.fn(key) % uint32(len(s.shards))
}

func defaultHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (s *sharder) Set(key string, value interface{}, expire time.Duration) (*Value, error) {
	i := s.getTargetShardIdx(key)
	if s.needLock {
		s.locks[i].Lock()
		defer s.locks[i].Unlock()
	}
	return s.shards[i].Set(key, value, expire)
}

func (s *sharder) Remove(key string) error {
	i := s.getTargetShardIdx(key)
	if s.needLock {
		s.locks[i].Lock()
		defer s.locks[i].Unlock()
	}
	return s.shards[i].Remove(key)
}

func (s *sharder) Get(key string) (*Value, error) {
	i := s.getTargetShardIdx(key)
	if s.needLock {
		s.locks[i].RLock()
		defer s.locks[i].RUnlock()
	}
	return s.shards[i].Get(key)
}

func (s *sharder) GetAtIndex(key string, subkey interface{}) (interface{}, error) {
	i := s.getTargetShardIdx(key)
	if s.needLock {
		s.locks[i].RLock()
		defer s.locks[i].RUnlock()
	}
	return s.shards[i].GetAtIndex(key, subkey)
}
func (s *sharder) Keys() ([]string, error) {
	n := len(s.shards) - 1
	resultCh := make(chan []string, n+1)
	errorCh := make(chan error, n+1)

	var wg sync.WaitGroup
	for idx := 0; idx < len(s.shards); idx++ {
		wg.Add(1)
		go func(i int) {
			if s.needLock {
				s.locks[i].RLock()
			}
			keys, err := s.shards[i].Keys()
			if s.needLock {
				s.locks[i].RUnlock()
			}

			resultCh <- keys
			errorCh <- err
			wg.Done()
		}(idx)
	}
	wg.Wait()
	close(errorCh)
	close(resultCh)

	for err := range errorCh {
		if err != nil {
			return []string{}, err
		}
	}

	result := []string{}
	for chunk := range resultCh {
		for i := range chunk {
			result = append(result, chunk[i])
		}
	}
	return result, nil
}
