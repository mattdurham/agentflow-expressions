package main

import (
	"agentflow/pkg/gragent"
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := interruptContext()
	defer cancel()
	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	s := gragent.NewSystem(l, "/Users/mdurham/Utils/agent_flow_configs/agent_expression_simply.hcl")

	if err := s.Load(); err != nil {
		panic(fmt.Errorf("error during the initial gragent load: %w", err))
	}
	r := mux.NewRouter()
	r.Handle("/graph", s.GraphHandler())
	// Gragent
	go func() {
		defer cancel()
		if err := s.Run(ctx); err != nil {
			level.Error(l).Log("msg", "error while running gragent", "err", err)
		}
	}()
	http.ListenAndServe(":54321", r)
}

func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
		case <-ctx.Done():
		}
		signal.Stop(sig)

		fmt.Fprintln(os.Stderr, "interrupt received")
	}()

	return ctx, cancel
}
