package db

import (
	"io"
	"log"
	"os"
	"time"
)

type logger struct {
	Cache
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newLogger(target Cache, writer io.Writer) *logger {
	if writer == nil {
		writer = os.Stderr
	}

	l := &logger{
		target,
		log.New(writer, "INFO: ", log.LstdFlags),
		log.New(writer, "ERROR: ", log.LstdFlags),
	}
	return l
}

func (l *logger) peekIntoPanic(params ...interface{}) {
	if r := recover(); r != nil {
		l.errorLog.Printf("PANIC in %v with error %v", params, r)
		panic(r)
	}
}

func (l *logger) Set(key string, value interface{}, expire time.Duration) (*Value, error) {
	defer l.peekIntoPanic("set", key, value, expire)
	result, err := l.Cache.Set(key, value, expire)
	l.infoLog.Println("set", key, value, expire, "=>", result, err)
	return result, err
}
func (l *logger) Get(key string) (*Value, error) {
	defer l.peekIntoPanic("get", key)
	result, err := l.Cache.Get(key)
	l.infoLog.Println("get", key, "=>", result, err)
	return result, err
}
func (l *logger) Remove(key string) error {
	defer l.peekIntoPanic("remove", key)
	err := l.Cache.Remove(key)
	l.infoLog.Println("remove", key, "=>", err)
	return err
}
func (l *logger) Keys() ([]string, error) {
	defer l.peekIntoPanic("keys")
	keys, err := l.Cache.Keys()
	l.infoLog.Println("keys", " => ", keys, err)
	return keys, err
}
func (l *logger) GetAtIndex(key string, index interface{}) (interface{}, error) {
	defer l.peekIntoPanic("getAtIndex", key, index)
	result, err := l.Cache.GetAtIndex(key, index)
	l.infoLog.Println("getAtIndex", key, index, " = >", result, err)
	return result, err

}
