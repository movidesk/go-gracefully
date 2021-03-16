package gracefully

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type GraceSuite struct {
	suite.Suite
}

type shutdownFake struct {
	s time.Duration
}

func newShutdown(s time.Duration) Shutdown {
	return &shutdownFake{s}
}

func (f *shutdownFake) Shutdown(ctx context.Context) error {
	time.Sleep(f.s)
	return nil
}

func (s *GraceSuite) TestShutdownWithTimeout() {
	assert := s.Assert()

	kill := make(chan os.Signal, 1)
	g := New(WithSignaler(kill), WithTimeout(time.Millisecond*100), WithShutdown(newShutdown(time.Second)))

	time.AfterFunc(time.Millisecond*100, func() { kill <- os.Interrupt })
	err := g.Grace()

	assert.EqualError(err, "closed by timeout")
}

func (s *GraceSuite) TestShutdownGracefully() {
	assert := s.Assert()

	kill := make(chan os.Signal, 1)
	g := New(WithSignaler(kill), WithTimeout(time.Second*1), WithShutdown(newShutdown(time.Millisecond*100)))

	time.AfterFunc(time.Millisecond*100, func() { kill <- os.Interrupt })
	err := g.Grace()

	assert.NoError(err)
}

func (s *GraceSuite) TestMultiShutdownWithTimeout() {
	assert := s.Assert()

	kill := make(chan os.Signal, 1)
	g := New(
		WithSignaler(kill),
		WithTimeout(time.Millisecond*100),
		WithShutdown(newShutdown(time.Second)),
		WithShutdown(newShutdown(time.Second)),
	)

	time.AfterFunc(time.Millisecond*100, func() { kill <- os.Interrupt })
	err := g.Grace()

	assert.EqualError(err, "closed by timeout")
}

func (s *GraceSuite) TestMultiShutdownGracefully() {
	assert := s.Assert()

	kill := make(chan os.Signal, 1)
	g := New(
		WithSignaler(kill),
		WithTimeout(time.Second*1),
		WithShutdown(newShutdown(time.Millisecond*100)),
		WithShutdown(newShutdown(time.Millisecond*100)),
	)

	time.AfterFunc(time.Millisecond*100, func() { kill <- os.Interrupt })
	err := g.Grace()

	assert.NoError(err)
}

func (s *GraceSuite) TestMixedShutdownWithTimeout() {
	assert := s.Assert()

	kill := make(chan os.Signal, 1)
	g := New(
		WithSignaler(kill),
		WithTimeout(time.Millisecond*200),
		WithShutdown(newShutdown(time.Second)),
		WithShutdown(newShutdown(time.Millisecond*200)),
	)

	time.AfterFunc(time.Millisecond*100, func() { kill <- os.Interrupt })
	err := g.Grace()

	assert.EqualError(err, "closed by timeout")
}

func TestGraceSuite(t *testing.T) {
	suite.Run(t, new(GraceSuite))
}
