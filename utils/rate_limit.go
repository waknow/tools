package utils

// const limit = 10000
// const request = 10000000
// func main() {
// 	rateLimit := NewRateLimit(limit, time.Second*10)
//
// 	wg := new(sync.WaitGroup)
//
// 	start := time.Now()
// 	for i := 0; i < 20; i++ {
// 		ip := fmt.Sprintf("ip.%d", i)
// 		go func() {
// 			defer wg.Done()
// 			start := time.Now()
// 			for i := 0; i < request; i++ {
// 				rateLimit.Check(ip)
// 			}
// 			diff := time.Now().Sub(start)
// 			log.Printf("%s -> %s %f", ip, diff, request/diff.Seconds())
// 		}()
// 		wg.Add(1)
// 	}
//
// 	wg.Wait()
// 	log.Println("all ->", time.Now().Sub(start))
// }

import (
	"container/list"
	"sync"
	"time"
)

type RateLimit struct {
	size      uint
	mutex     *sync.RWMutex
	recordMap map[string]*Record
	limit     uint
	expire    time.Duration
}

func NewRateLimit(limit uint, expire time.Duration) *RateLimit {
	return &RateLimit{
		mutex:     new(sync.RWMutex),
		recordMap: make(map[string]*Record),
		limit:     limit,
		expire:    expire,
	}
}

func (i *RateLimit) Init() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C
			i.clean()
		}
	}()
}

func (i *RateLimit) Change(limit uint, expire time.Duration) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.limit = limit
	i.expire = expire
}

func (i *RateLimit) Check(ip string) bool {
	i.mutex.RLock()
	record, ok := i.recordMap[ip]
	i.mutex.RUnlock()
	if !ok {
		i.mutex.Lock()
		record = NewRecord()
		record.Init()
		i.recordMap[ip] = record
		i.size++
		i.mutex.Unlock()
	}
	return record.Check(i.limit, i.expire)
}

func (i *RateLimit) Remove(ip string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if _, ok := i.recordMap[ip]; ok {
		delete(i.recordMap, ip)
		i.size--
	}
}

func (i *RateLimit) Size() uint {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.size
}

func (i *RateLimit) clean() {
	i.mutex.RLock()
	ipNeedRemove := []string{}
	for ip, record := range i.recordMap {
		record.RemoveExpire(i.expire)
		if record.Number() == 0 {
			ipNeedRemove = append(ipNeedRemove, ip)
		}
	}
	i.mutex.RUnlock()

	for _, ip := range ipNeedRemove {
		i.Remove(ip)
	}
}

type Record struct {
	mutex  *sync.RWMutex
	list   *list.List
	number uint
}

func NewRecord() *Record {
	return &Record{
		mutex: new(sync.RWMutex),
		list:  list.New(),
	}
}

func (i *Record) Init() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.list.Init()
}

func (i *Record) Check(limit uint, expire time.Duration) bool {
	if i.Number() >= limit {
		return true
	} else {
		i.Add()
	}
	i.RemoveExpire(expire)
	return false
}

func (i *Record) Number() uint {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.number
}

func (i *Record) Add() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.list.PushBack(time.Now().UnixNano())
	i.number++
}

func (i *Record) RemoveExpire(dur time.Duration) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	current := time.Now().UnixNano()
	diff := dur.Nanoseconds()
	front := i.list.Front()
	for front != nil {
		if current-front.Value.(int64) >= diff {
			t := front
			front = front.Next()
			i.list.Remove(t)
			i.number--
		} else {
			break
		}
	}
}
