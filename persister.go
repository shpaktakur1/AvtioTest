package db

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

var ErrUnknownOperationType = errors.New("Unknown operation type")

type persister struct {
	Cache

	oplog chan operation
	oplog []operation

	rw io.ReadWriter

	sync.RWMutex
}

type operation struct {
	Type   string      `json:"Type"`
	Key    string      `json:"k"`
	Value  interface{} `json:"v"`
	Expire int64       `json:"e"`
}

func (p *persister) restore(source io.Reader) error {
	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		op := operation{}
		err := json.Unmarshal(scanner.Bytes(), &op)
		if err != nil {
			return err
		}

		go func() { <-p.op }()
		err = op.execute(p)

		if err != nil && err != ErrInvalidTTL {
			return err
		}
	}
	return nil
}

// работает только при запуске
func (o *operation) execute(target Cache) (err error) {
	switch o.Type {
	case "Set":
		if o.Expire != 0 {
			nowNano := time.Now().UnixNano()
			if nowNano >= o.Expire {
				target.Remove(o.Key)
				err = ErrInvalidTTL
				return
			} else {
				o.Expire = o.Expire - nowNano
			}
		}
		_, err = target.Set(o.Key, o.Value, time.Duration(o.Expire))
	case "Remove":
		err = target.Remove(o.Key)
	default:
		err = ErrUnknownOperationType
	}
	return
}

func (p *persister) consumeOplog() {
	for message := range p.op {
		p.RWMutex.Lock()
		p.oplog = append(p.oplog, message)
		p.RWMutex.Unlock()
	}
}

func (p *persister) writeOplogEvery(frequency time.Duration) {
	for {
		<-time.After(frequency)
		p.writeOplog(p.grabOplog())
	}

}

func (p *persister) grabOplog() []operation {
	p.RWMutex.RLock()
	defer p.RWMutex.RUnlock()
	ops := p.oplog
	p.oplog = []operation{}
	return ops
}
func (p *persister) writeOplog(ops []operation) {
	if len(ops) > 0 {
		for i := range ops {
			entry, err := json.Marshal(ops[i])
			if err != nil {
				log.Fatal(err, ops[i]) //операция из БД была успешной для необработанных данных
			}
			p.rw.Write(entry)
			p.rw.Write([]byte("\n"))
		}
	}
}

func newPersister(target Cache, srcDst io.ReadWriter, writeFrequency time.Duration) (*persister, error) {
	p := &persister{
		Cache: target,
		op:    make(chan operation),
		oplog: []operation{},
		rw:    srcDst,
	}

	if srcDst != nil {
		err := p.restore(srcDst)
		if err != nil {
			return nil, err
		}
	}

	go p.consumeOplog()

	if srcDst != nil {
		go p.writeOplogEvery(writeFrequency)
	}
	return p, nil
}

func (p *persister) Set(key string, value interface{}, expire time.Duration) (*Value, error) {
	result, err := p.Cache.Set(key, value, expire)
	if err == nil {
		p.op <- operation{"Set", key, value, result.Expires}
	}
	return result, err
}

func (p *persister) Remove(key string) error {
	err := p.Cache.Remove(key)
	if err == nil {
		p.op <- operation{"Remove", key, nil, 0}
	}
	return err
}
