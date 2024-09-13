/*******************************************************************************
 * Copyright 2021 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
/*
	Package bootstrap contains all abstractions and implementation necessary to bootstrap the application.
*/
package bootstrap

import (
	"context"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Run is the bootstrap process entry point. All relevant application components can be initialized by providing a
// BootstrapHandler implementation to the handlers array.
func Run(
	ctx context.Context,
	cancel context.CancelFunc,
	configuration config.Configuration,
	handlers []BootstrapHandler) {

	wg, _ := initWaitGroup(ctx, cancel, configuration, handlers)   // main entry point for bootstrap handlers/process

	wg.Wait()             // waits until all tasks are done
}

func initWaitGroup(
	ctx context.Context,
	cancel context.CancelFunc,
	configuration config.Configuration,
	handlers []BootstrapHandler) (*sync.WaitGroup, bool) {

	startedSuccessfully := true

	var wg sync.WaitGroup              // used to keep track of asynchronous task
	// call individual bootstrap handlers.
	translateInterruptToCancel(ctx, &wg, cancel)  //ensure that a SIGTERM signal (graceful termination request) will cancel the context and allow for clean shutdowns
	for i := range handlers {
		if handlers[i](ctx, &wg) == false {
			cancel()
			startedSuccessfully = false
			break
		}
	}

	return &wg, startedSuccessfully
}

// translateInterruptToCancel spawns a go routine to translate the receipt of a SIGTERM signal to a call to cancel
// the context used by the bootstrap implementation.
func translateInterruptToCancel(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		signalStream := make(chan os.Signal)   //creates channel for go routines to communicate with each other, send signsls like SIGINT (interrupt) or SIGTERM (terminate).
		defer func() {
			signal.Stop(signalStream)
			close(signalStream)
		}()
		// It tells Go's signal handling system to listen for os.Interrupt(Cltr+C) and syscall.SIGTERM signals and send them to the signalStream channel.
		signal.Notify(signalStream, os.Interrupt, syscall.SIGTERM)   //listens for system signals and maps them to cancellation behavior. If a signal is received, it triggers the cancel()
		select {
		case <-signalStream:     //A system interrupt signal (signalstream received any signal)
			cancel()
			return
		case <-ctx.Done():     //Context Cancellation (ctx.Done()), which indicates that the program is already stopping,
			return
		}
	}()
}
