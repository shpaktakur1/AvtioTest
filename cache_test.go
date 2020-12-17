package db

import (
	"strconv"
	"testing"
	"time"
)

func getRandomLists(n int, l int) map[string][]int {
	items := make(map[string][]int)
	for i := 0; i < n; i++ {
		key := strconv.Itoa(i)
		items[key] = make([]int, l)
		for j := 0; j < l; j++ {
			items[key][j] = j
		}
	}
	return items
}
func TestCache_Concurrency(t *testing.T) {
	c, err := NewCache(5*time.Millisecond, nil, nil, 0, 0, nil)
	if err != nil {
		t.Error(err)
	}
	items := getRandomLists(500, 10)
	done := make(chan bool)

	for a := 0; a < 100; a++ {
		for k, v := range items {
			go func(key string, value []int) {
				c.Set(key, value, 0)
				done <- true
			}(k, v)
			go func(key string) {
				c.Get(key)
				done <- true
			}(k)
			go func(key string) {
				c.Remove(key)
				done <- true
			}(k)
			go func(key string) {
				for i := 0; i < 10; i++ {
					c.GetAtIndex(key, i)
				}
				done <- true
			}(k)
		}
	}

	for i := 0; i < len(items)*4*100; i++ {
		<-done
	}

}
