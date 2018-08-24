package future

// Future
type Future struct {
	id string
	ch chan error
}

// 创建Future
func newFuture(id string) *Future {
	ch := make(chan error)
	return &Future{ch: ch, id: id}
}

// 获取ID
func (f *Future) ID() string {
	return f.id
}

// 获取结果
func (f *Future) GetResult() error {
	err := <-f.ch
	return err
}

// 设置结果
func (f *Future) SetResult(err error) {
	f.ch <- err
}
