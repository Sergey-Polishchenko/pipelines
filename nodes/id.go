package nodes

import "sync/atomic"

var globalNodeCounter atomic.Uint64

func nextNodeID() uint64 {
	return globalNodeCounter.Add(1)
}
