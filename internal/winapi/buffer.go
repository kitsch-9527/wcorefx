//go:build windows

package winapi

import (
	"errors"
)

const (
	defaultBufSize = 4096
	maxRetries     = 5
)

// BufferStrategy is the interface for a single buffer-fill attempt.
// Fill returns the number of bytes written and an error.
// If the error is ErrInsufficientBuffer, BufferQuery grows and retries.
type BufferStrategy interface {
	Fill([]byte) (int, error)
}

// FuncStrategy adapts a function to BufferStrategy.
type FuncStrategy func([]byte) (int, error)

func (f FuncStrategy) Fill(buf []byte) (int, error) { return f(buf) }

// BufferQuery runs the two-call buffer-size-negotiation loop automatically.
//
// It calls s.Fill with a growing buffer until Fill returns nil or a
// non-ErrInsufficientBuffer error.  At most maxRetries attempts are made.
func BufferQuery(s BufferStrategy) ([]byte, error) {
	buf := make([]byte, defaultBufSize)
	for i := 0; i < maxRetries; i++ {
		used, err := s.Fill(buf)
		if err == nil {
			return buf[:used], nil
		}
		var ib *ErrInsufficientBuffer
		if errors.As(err, &ib) {
			if ib.Size > len(buf) {
				buf = make([]byte, ib.Size)
			} else {
				buf = make([]byte, len(buf)*2)
			}
			continue
		}
		return nil, err
	}
	return nil, errors.New("buffer query: max retries exceeded")
}
