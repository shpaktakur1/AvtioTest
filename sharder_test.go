package db

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getRandomStrings(n int) []string {
	result := make([]string, n)
	for i := 0; i < n; i++ {
		b := make([]byte, n)
		for j := range b {
			b[j] = letters[rand.Intn(len(letters))]
		}
		result[i] = string(b)
	}
	return result
}

func TestShard_Keys(t *testing.T) {
	var tests = []int{1, 2, 3, 7, 9}
	keys := getRandomStrings(50)
	keysMap := make(map[string]bool)
	for i := range keys {
		keysMap[keys[i]] = true
	}
	uniqueKeys := make([]string, len(keysMap))

	i := 0
	for k := range keysMap {
		uniqueKeys[i] = k
		i++
	}

	sort.Strings(uniqueKeys)
	for _, n := range tests {
		sharder, _ := newSharder(n, nil)
		done := make(chan bool, len(keys))
		for i := range keys {
			go func(idx int) {
				sharder.Set(keys[idx], keys[idx], 0)
				done <- true
			}(i)
		}
		for i := 0; i < len(keys); i++ {
			<-done
		}
		close(done)
		actualKeys, err := sharder.Keys()
		if err != nil {
			t.Error(err)
		}
		sort.Strings(actualKeys)
		if !reflect.DeepEqual(actualKeys, uniqueKeys) {
			t.Error("keys and actual Keys do not match", actualKeys, uniqueKeys)
		}
	}
}
