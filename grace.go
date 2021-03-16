package gracefully

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Shutdown interface {
	Shutdown(context.Context) error
}

type GraceFn func(*grace)

func WithTimeout(t time.Duration) GraceFn {
	return func(g *grace) {
		g.td = t
	}
}

func WithShutdown(s Shutdown) GraceFn {
	return func(g *grace) {
		g.ss = append(g.ss, s)
	}
}

func WithSignaler(sig chan os.Signal) GraceFn {
	return func(g *grace) {
		g.sig = sig
	}
}

type Grace interface {
	Grace() error
}

type grace struct {
	ss  []Shutdown
	td  time.Duration
	sig chan os.Signal
}

func New(fns ...GraceFn) Grace {
	g := &grace{}
	WithTimeout(time.Second * 5)(g)
	WithSignaler(make(chan os.Signal, 1))(g)
	for _, fn := range fns {
		fn(g)
	}
	return g
}

func (g *grace) Grace() error {
	ctx, cancel := context.WithTimeout(context.Background(), g.td)
	defer cancel()
	timeout := ctx.Done()

	signal.Notify(g.sig, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	<-g.sig

	done := make(chan struct{})

	go g.shutdownAll(ctx, done)

	select {
	case <-done:
		return nil
	case <-timeout:
		return errors.New("closed by timeout")
	}
}

func (g *grace) shutdownAll(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, s := range g.ss {
			wg.Add(1)
			go g.shutdownOne(ctx, wg, s)
		}
	}()
	wg.Wait()
}

func (g *grace) shutdownOne(ctx context.Context, wg *sync.WaitGroup, s Shutdown) {
	defer wg.Done()
	if err := s.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
