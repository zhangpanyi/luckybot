package monitor

// 过期信息
type expire struct {
	ID        uint64 // 红包ID
	Timestamp int64  // 时间戳
}

// 堆结构
type heapExpire []expire

// 堆大小
func (h heapExpire) Len() int { return len(h) }

// 比较大小
func (h heapExpire) Less(i, j int) bool {
	if h[i].Timestamp == h[j].Timestamp {
		return h[i].ID < h[j].ID
	}
	return h[i].Timestamp < h[j].Timestamp
}

// 交换元素
func (h heapExpire) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// 添加元素
func (h *heapExpire) Push(x interface{}) {
	*h = append(*h, x.(expire))
}

// 删除元素
func (h *heapExpire) Pop() interface{} {
	if h.Len() == 0 {
		return nil
	}
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// 首个元素
func (h *heapExpire) Front() *expire {
	if h.Len() == 0 {
		return nil
	}
	return &(*h)[0]
}
