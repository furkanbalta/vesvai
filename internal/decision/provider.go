package decision

import "context"

type ProviderError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ProviderError) Error() string {
	return e.Message
}

func (e *ProviderError) Temporary() bool {
	return e.StatusCode == 429 || e.StatusCode >= 500
}

type Provider interface {
	Name() string
	Decide(ctx context.Context, req *Request) (*Response, error)
}
