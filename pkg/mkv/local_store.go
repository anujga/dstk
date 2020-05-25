package mkv

type InsertOptions int
type GetOptions int

const (
	//behaviour mask = 0x3
	FailOnExisting InsertOptions = 1
	OverwriteExisting InsertOptions = 2


	//consistency mask = 0x3 << 2
	FlushImmediately InsertOptions = 4 + 0
)

type StoreEntry struct {
	key, value []byte
	metadata uint64
	//ttl: this store may implement a very crude form of ttl during compaction
	//https://github.com/facebook/rocksdb/wiki/Merge-Operator
}

type LocalStore interface {
	Insert(entries []StoreEntry) int64
	Get(keys []byte, options GetOptions) StoreEntry
	Remove(key []byte)
}


