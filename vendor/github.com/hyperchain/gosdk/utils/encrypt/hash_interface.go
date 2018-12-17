//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package encrypt

import "github.com/hyperchain/gosdk/common"

// hash interface
type CommonHash interface {
	Hash(x interface{}) (h common.Hash)
	ByteHash(data ...[]byte) (h common.Hash)
}
