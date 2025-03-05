package reverse_index

import (
	"github.com/huandu/skiplist"
	farmhash "github.com/leemcloughlin/gofarmhash"
	"runtime"
	"sync"
	"zcygo/types"
	"zcygo/util"
)

type SkiplistReverseIndex struct {
	table *util.ConcurrentHashMap
	locks []sync.RWMutex
}

// NewSkipListReverseIndex 创建并初始化一个新的跳表倒排索引实例。
// DocNumEstimate 参数用于估计文档数量，以优化并发哈希映射的性能。
func NewSkipListReverseIndex(DocNumEstimate int) *SkiplistReverseIndex {
	// 创建一个新的SkiplistReverseIndex实例。
	indexer := new(SkiplistReverseIndex)

	// 初始化并发哈希映射，用于存储跳表倒排索引的数据。
	// 使用runtime.NumCPU()来确定哈希映射的并发度，以优化并发性能。
	indexer.table = util.NewConcurrentHashMap(runtime.NumCPU(), DocNumEstimate)

	// 初始化读写锁数组，用于在并发访问时保护跳表倒排索引的数据完整性。
	// 这里的1000是预设的锁的数量，可以根据实际需要进行调整。
	indexer.locks = make([]sync.RWMutex, 1000)

	// 返回初始化后的SkiplistReverseIndex实例。
	return indexer
}

func (indexer SkiplistReverseIndex) getLock(key string) *sync.RWMutex {
	n := int(farmhash.Hash32WithSeed([]byte(key), 0))
	return &indexer.locks[n%len(indexer.locks)]
}

type SkipListValue struct {
	Id          string
	BitsFeature uint64
}

// Add 方法用于向倒排索引中添加文档信息。
// 它遍历文档的关键词，并为每个关键词在跳表中添加相应的文档信息。
// 如果关键词已存在于索引中，则更新跳表中的文档信息；
// 如果关键词不存在，则创建新的跳表并添加文档信息。
func (indexer *SkiplistReverseIndex) Add(doc *types.Document) {
	// 遍历文档的所有关键词
	for _, keyword := range doc.Keywords {
		// 将关键词转换为字符串形式
		key := keyword.ToString()
		lock := indexer.getLock(key)
		lock.Lock()
		// 创建跳表值对象，包含文档ID和位特征
		sklValue := SkipListValue{
			Id:          doc.Id,
			BitsFeature: doc.BitsFeature,
		}

		// 检查关键词是否已存在于索引中
		if value, ok := indexer.table.Get(key); ok {
			// 如果存在，则获取对应的跳表并添加文档信息
			list := value.(*skiplist.SkipList)
			list.Set(doc.IntId, sklValue)
		} else {
			// 如果不存在，则创建新的跳表，并添加文档信息
			list := skiplist.New(skiplist.Uint64)
			list.Set(doc.IntId, sklValue)
			indexer.table.Set(key, list)
		}
		lock.Unlock()
	}
}

// Delete 从跳表逆向索引中删除指定的关键词与ID的关联。
// 此函数主要用于移除与给定关键词相关联的特定ID。
// 参数:
//   - IntId: 要从索引中删除的记录的ID。
//   - keyword: 指示要从索引中删除的关键词。
func (indexer *SkiplistReverseIndex) Delete(IntId uint64, keyword *types.Keyword) {
	// 将关键词转换为字符串形式，以便在索引表中进行查找。
	key := keyword.ToString()

	// 获取与该关键词相关的锁，以确保并发安全性。
	lock := indexer.getLock(key)

	// 锁定，以安全地访问和修改索引表。
	lock.Lock()
	defer lock.Unlock() // 使用defer确保函数退出时解锁。

	// 尝试从索引表中获取与关键词相关联的跳表。
	if value, ok := indexer.table.Get(key); ok {
		// 将获取到的值转换为跳表类型。
		list := value.(*skiplist.SkipList)

		// 从跳表中移除指定的ID。
		list.Remove(IntId)
	}
}
