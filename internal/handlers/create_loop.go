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
	"encoding/json"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
	"github.com/project-alvarium/example-go/internal/models"
	"log/slog"
	"sync"
	"time"
)

type CreateLoop struct {
	cfg       config.SdkInfo
	chPublish chan []byte
	logger    interfaces.Logger
	sdk       interfaces.Sdk
}

func NewCreateLoop(sdk interfaces.Sdk, ch chan []byte, cfg config.SdkInfo, logger interfaces.Logger) CreateLoop {
	return CreateLoop{
		cfg:       cfg,
		chPublish: ch,
		logger:    logger,
		sdk:       sdk,
	}
}

func (c *CreateLoop) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup) bool {
	cancelled := false
	wg.Add(1)
	go func() {
		defer wg.Done()

		for !cancelled {
			data, err := models.NewSampleData(c.cfg.Signature.PrivateKey)
			if err != nil {
				c.logger.Error(err.Error())
				continue
			}
			b, _ := json.Marshal(data)

			c.sdk.Create(context.Background(), b)
			c.chPublish <- b
			time.Sleep(1 * time.Second)
		}
		close(c.chPublish)
		c.logger.Write(slog.LevelDebug, "cancel received")
	}()

	wg.Add(1)
	go func() { // Graceful shutdown
		defer wg.Done()

		<-ctx.Done()
		c.logger.Write(slog.LevelInfo, "shutdown received")
		cancelled = true
	}()
	return true
}
