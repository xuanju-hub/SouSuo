package util

import (
	"github.com/leemcloughlin/gofarmhash"
	"maps"
	"slices"
	"sync"
)

//自行实现支持并发读写的map，key是string，value是任意类型

type ConcurrentHashMap struct {
	maps  []map[string]any // 由多个小map组成
	seg   int              // 小map的数量
	locks []sync.RWMutex   // 每个小map配一把读写锁。避免全局只有一把锁。影响性能
	seed  uint32           // 每次执行farmhash的种子
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
func (m *ConcurrentHashMap) CreateIterator() *ConcurrentHashMapIterator {
	keys := make([][]string, 0, len(m.maps))
	for _, mp := range m.maps {
		row := slices.Sorted(maps.Keys(mp))
		keys = append(keys, row)
	}

	return &ConcurrentHashMapIterator{
		cm:       m,
		keys:     keys,
		rowIndex: 0,
		colIndex: 0,
	}
}

type MapEntry struct {
	Key   string
	Value any
}

type MapIterator interface {
	Next() *MapEntry
}

type ConcurrentHashMapIterator struct {
	cm       *ConcurrentHashMap
	keys     [][]string
	rowIndex int
	colIndex int
}

// Next 返回迭代器的下一个 MapEntry 对象。
// 如果迭代器已经遍历完所有的键值对，则返回 nil。
func (iter *ConcurrentHashMapIterator) Next() *MapEntry {
	// 检查当前行索引是否超出 keys 数组的范围。
	if iter.rowIndex >= len(iter.keys) {
		return nil
	}

	// 获取当前行的键集合。
	row := iter.keys[iter.rowIndex]
	// 如果当前行为空，则递归调用 Next 方法，直到找到非空行。
	if len(row) == 0 { //本行为空
		iter.rowIndex++
		return iter.Next() // 继续往下找
	}

	// 获取本行的第colIndex个key，并根据该key获取对应的value。
	key := row[iter.colIndex] // 获取本行的第colIndex个key
	value, _ := iter.cm.Get(key)

	// 更新列索引，如果当前列是最后一列，则更新行索引并重置列索引。
	if iter.colIndex >= len(row)-1 {
		iter.rowIndex++
		iter.colIndex = 0
	} else {
		iter.colIndex++
	}

	// 返回当前键值对的 MapEntry 对象。
	return &MapEntry{
		Key:   key,
		Value: value,
	}
}
