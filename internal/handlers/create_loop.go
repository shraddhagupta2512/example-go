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

//sets up a loop that continuously generates and processes new sample data, annotates it, and publishes it to a stream, 
//also handles graceful shutdown when a cancellation signal is received
// Parameters: ctx context.Context: used for cancellation and coordination of the process. 
//wg *sync.WaitGroup: ensures all concurrent goroutines are properly synchronized and wait for each other to finish before exiting.
func (c *CreateLoop) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup) bool {
	cancelled := false       //keep track of when the process should stop during graceful shutdown
	wg.Add(1)
	go func() {
		defer wg.Done()

		for !cancelled {
			data, err := models.NewSampleData(c.cfg.Signature.PrivateKey)   //data is generated using the private key using NewSampleData function
			if err != nil {
				c.logger.Error(err.Error())
				continue
			}
			b, _ := json.Marshal(data)    //data is converted into json format

			c.sdk.Create(context.Background(), b)    //send data to the sdk.create, which annotates the data using the SDK's configured annotators
			c.chPublish <- b                        //raw data is published to channel
			time.Sleep(1 * time.Second)             //waits befoe generating new data
		}
		close(c.chPublish)                          //close the channel (cancel is true; no more data will be generated/sent)
		c.logger.Write(slog.LevelDebug, "cancel received")
	}()

	wg.Add(1)
	go func() { // Graceful shutdown
		defer wg.Done()

		<-ctx.Done()                      //cancellation signal from context (main application is shutting down) 
		c.logger.Write(slog.LevelInfo, "shutdown received")
		cancelled = true
	}()
	return true                     //handler started successfully
}
