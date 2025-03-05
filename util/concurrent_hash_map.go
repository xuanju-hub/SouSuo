package util

import (
	"sync"
	"github.com/leemcloughlin/gofarmhash"
)

//自行实现支持并发读写的map，key是string，value是任意类型

type ConcurrentHashMap struct {
	maps  []map[string]any // 由多个小map组成
	seg   int              // 小map的数量
	locks []sync.RWMutex   // 每个小map配一把读写锁。避免全局只有一把锁。影响性能
	seed  uint32             // 每次执行farmhash的种子
}

//cap预估map中容纳多少元素，seg内部包含几个小map

func NewConcurrentHashMap(cap, seg int) *ConcurrentHashMap {
	maps := make([]map[string]any, seg)
	locks := make([]sync.RWMutex, seg)
	for i := 0; i < seg; i++ {
		maps[i] = make(map[string]any, cap/seg)
	}
	return &ConcurrentHashMap{
		maps:  maps,
		seg:   seg,
		locks: locks,
		seed:  0,
	}
}
// 判断key对应到哪个小map
func (m *ConcurrentHashMap) getSegIndex(key string) int {
	hash := int(farmhash.Hash32WithSeed([]byte(key), m.seed))
	return hash % m.seg
}

// 写入<key, value>，如果key已经存在，则覆盖value
func (m *ConcurrentHashMap) Set(key string, value any) {
	index := m.getSegIndex(key)
	m.locks[index].Lock()
	m.maps[index][key] = value
	m.locks[index].Unlock()
}

// 根据key获取value

func (m *ConcurrentHashMap) Get(key string) (value any, ok bool) {
	index := m.getSegIndex(key)
	m.locks[index].RLock()
	value, ok = m.maps[index][key]
	m.locks[index].RUnlock()
	return
}

