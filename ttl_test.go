package db

import (
	"testing"
	"time"
)

func TestTTL_Negative(t *testing.T) {
	s := newStore()
	_, err := newTtl(s, -1)
	if err == nil || err != ErrInvalidTTL {
		t.Errorf("TestTTL_Negative got no error %v", ErrInvalidTTL)
	}

}
func TestTTL_NoDefault(t *testing.T) {
	s := newStore()
	ttl, err := newTtl(s, 0)
	if err != nil {
		t.Errorf("TestTTL_NoDefault got error %v", err)
	}
	ttl.Set("key_1", "value", 1*time.Millisecond)
	ttl.Set("key_default", "value", 0)
	item, err := ttl.Get("key_1")
	if err != nil || item.Data != "value" {
		t.Errorf("TestTTL_Nodefault failed to Get fresh value %v, err:%v", item, err)
	}
	time.Sleep(time.Millisecond * 2)
	item, err = ttl.Get("key_1")
	if err == nil || item != nil {
		t.Errorf("TestTTL_Nodefault got expired value %v, err:%v", item, err)
	}
	item, err = ttl.Get("key_default")
	if err != nil || item == nil {
		t.Errorf("TestTTL_Nodefault - value with no expiration date was expired")
	}

	_, err = ttl.Set("key_default", "value", -1)
	if err == nil || err != ErrInvalidTTL {
		t.Errorf("TestTTL got no error for negative TTL")
	}

}

func TestTTL_Default(t *testing.T) {
	s := newStore()
	ttl, err := newTtl(s, 1*time.Nanosecond)
	if err != nil {
		t.Errorf("TestTTL_NoDefault got error %v", err)
	}
	ttl.Set("key_1", "value", 1*time.Millisecond)
	ttl.Set("key_default", "value", 0)
	item, err := ttl.Get("key_1")
	if err != nil || item.Data != "value" || item.Expires == 0 {
		t.Errorf("TestTTL_Default failed to Get fresh value %v, err:%v", item, err)
	}
	time.Sleep(time.Millisecond * 2)
	item, err = ttl.Get("key_1")
	if err == nil || item != nil {
		t.Errorf("TestTTL_Default got expired value %v, err:%v", item, err)
	}
	item, err = ttl.Get("key_default")
	if err == nil || item != nil {
		t.Errorf("TestTTL_Default - value with no expiration date was not expired by default")
	}
}
