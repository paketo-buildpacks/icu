package fakes

import "sync"

type LayerArranger struct {
	ArrangeCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Error error
		}
		Stub func(string) error
	}
}

func (f *LayerArranger) Arrange(param1 string) error {
	f.ArrangeCall.mutex.Lock()
	defer f.ArrangeCall.mutex.Unlock()
	f.ArrangeCall.CallCount++
	f.ArrangeCall.Receives.Path = param1
	if f.ArrangeCall.Stub != nil {
		return f.ArrangeCall.Stub(param1)
	}
	return f.ArrangeCall.Returns.Error
}
