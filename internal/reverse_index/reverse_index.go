package reverse_index

import (
	"zcygo/types"
)

type IReverseIndex interface {
	Add(doc types.Document)
	Delete(IntId uint64, keyword types.Keyword)
	//Search(q *types.TermQuery, onFlag uint64, offFlag uint64, orderFlag []uint64) []string
}
