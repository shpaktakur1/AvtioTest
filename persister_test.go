package db

import (
	"bytes"
	"testing"
	"time"
)

var sample = `{"Type":"Set","k":"foo","v":"bar","e":0}
{"Type":"Set","k":"test","v":[1,2,3],"e":0}
{"Type":"Remove","k":"test","v":null,"e":0}
`

func TestPersister_Write(t *testing.T) {
	store := newStore()
	rw := bytes.Buffer{}
	p, _ := newPersister(store, &rw, 1*time.Nanosecond)
	p.Set("foo", "bar", 0)
	p.Set("test", []interface{}{1, 2, 3}, 0)
	p.Remove("test")
	time.Sleep(5 * time.Millisecond)
	result := rw.String()
	if result != sample {
		t.Errorf("TestPersister_Oplog REMOVE expected:\n%v\ngot:\n%v", sample, result)
	}
}

func TestPersister_Read(t *testing.T) {
	store := newStore()
	rw := bytes.Buffer{}
	rw.Write([]byte(sample))
	p, err := newPersister(store, &rw, 1*time.Nanosecond)
	if err != nil {
		t.Errorf("TestPersister_Read got constructor error %v", err)
	}
	time.Sleep(1 * time.Millisecond)
	foo, err := p.Get("foo")
	if err != nil {
		t.Errorf("TestPersister_Read got unexpected error %v", err)
		return
	}
	if foo.Data != "bar" {
		t.Errorf("TestPersister_Read .Get(foo) expected bar, got %v", foo.Data)
	}
}

func TestPersister_ReadWrite(t *testing.T) {
	store := newStore()
	rw := bytes.Buffer{}
	rw.Write([]byte(sample))
	p, err := newPersister(store, &rw, 1*time.Nanosecond)
	if err != nil {
		t.Errorf("TestPersister_Read got constructor error %v", err)
	}
	time.Sleep(1 * time.Millisecond)
	p.Set("new_foo", "new_bar", 0)
	p.Set("new_test", []interface{}{1, 2, 3}, 0)
	time.Sleep(1 * time.Millisecond)
	expected := `{"Type":"Set","k":"new_foo","v":"new_bar","e":0}
{"Type":"Set","k":"new_test","v":[1,2,3],"e":0}
` // Здесь только новые значения
	if rw.String() != expected {
		t.Errorf("TestPersister_ReadWrite expected %v, \ngot %v", expected, rw.String())
	}

}
