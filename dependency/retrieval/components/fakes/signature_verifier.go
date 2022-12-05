package fakes

import "sync"

type SignatureVerifier struct {
	VerifyCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			SignatureURL string
			TargetURL    string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
}

func (f *SignatureVerifier) Verify(param1 string, param2 string) error {
	f.VerifyCall.mutex.Lock()
	defer f.VerifyCall.mutex.Unlock()
	f.VerifyCall.CallCount++
	f.VerifyCall.Receives.SignatureURL = param1
	f.VerifyCall.Receives.TargetURL = param2
	if f.VerifyCall.Stub != nil {
		return f.VerifyCall.Stub(param1, param2)
	}
	return f.VerifyCall.Returns.Error
}
