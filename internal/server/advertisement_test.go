package server

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
)

func TestAdvertisementStopWaitsForResponder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	advertisement := &Advertisement{
		cancel: cancel,
		done:   done,
		log:    hclog.NewNullLogger(),
	}

	responderFinished := make(chan struct{})
	go func() {
		<-ctx.Done()
		time.Sleep(25 * time.Millisecond)
		close(responderFinished)
		close(done)
	}()

	advertisement.Stop()
	select {
	case <-responderFinished:
	default:
		t.Fatal("Stop returned before the responder finished")
	}

	// Stop must remain safe if cleanup is requested more than once.
	advertisement.Stop()
}
