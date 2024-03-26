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
package handlers

import (
	"context"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
	"log/slog"
	"sync"
)

type Transit struct {
	cfg         config.SdkInfo
	chSubscribe chan []byte
	logger      interfaces.Logger
	sdk         interfaces.Sdk
}

func NewTransit(sdk interfaces.Sdk, chSub chan []byte, cfg config.SdkInfo, logger interfaces.Logger) Transit {
	return Transit{
		cfg:         cfg,
		chSubscribe: chSub,
		logger:      logger,
		sdk:         sdk,
	}
}

func (t *Transit) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup) bool {
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msg, ok := <-t.chSubscribe
			if ok {
				t.sdk.Transit(ctx, msg)
			} else { //channel has been closed. End goroutine.
				t.logger.Write(slog.LevelInfo, "transit::chSubscribe closed, exiting")
				return
			}
		}
	}()

	wg.Add(1)
	go func() { // Graceful shutdown
		defer wg.Done()

		<-ctx.Done()
		t.logger.Write(slog.LevelInfo, "shutdown received")
	}()
	return true
}
