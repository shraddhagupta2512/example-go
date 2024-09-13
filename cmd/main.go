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
package main

import (
	"context"
	"flag"
	"github.com/project-alvarium/alvarium-sdk-go/pkg"
	SdkConfig "github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/factories"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
	"github.com/project-alvarium/example-go/internal/bootstrap"
	"github.com/project-alvarium/example-go/internal/config"
	"github.com/project-alvarium/example-go/internal/handlers"
	"log/slog"
	"os"
	"fmt"
)

func main() {
	// Load config
	var configPath string
	flag.StringVar(&configPath,
		"cfg",
		"./res/config.json",
		"Path to JSON configuration file.")
	flag.Parse()

	fileFormat := config.GetFileExtension(configPath)
	reader, err := config.NewReader(fileFormat)
	if err != nil {
		tmpLog := factories.NewLogger(SdkConfig.LoggingInfo{MinLogLevel: slog.LevelError})
		tmpLog.Error(err.Error())
		os.Exit(1)
	}

	cfg := config.ApplicationConfig{}
	err = reader.Read(configPath, &cfg)
	if err != nil {
		tmpLog := factories.NewLogger(SdkConfig.LoggingInfo{MinLogLevel: slog.LevelError})
		tmpLog.Error(err.Error())
		os.Exit(1)
	}

	logger := factories.NewLogger(cfg.Logging)
	logger.Write(slog.LevelDebug, "config loaded successfully")
	logger.Write(slog.LevelDebug, cfg.AsString())

	// List of annotators driven from config, eventually support dist. policy.
	var annotators []interfaces.Annotator
	for _, t := range cfg.Sdk.Annotators {
		instance, err := factories.NewAnnotator(t, cfg.Sdk)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		annotators = append(annotators, instance)
	}
	// debugging list of annotators:
	for i, annotator := range annotators {
		fmt.Printf("Annotator %d: %+v\n", i+1, annotator)
	} 
	sdk := pkg.NewSdk(annotators, cfg.Sdk, logger)

	// print the contents of sdk
	fmt.Printf("******** SDK content*******: %+v\n", sdk)

	// print the contents of cfg.sdk
	fmt.Printf("******** CFG SDK content*******: %+v\n", cfg.Sdk)

	// handlers for example functionality
	chCreate := make(chan []byte)
	chMutate := make(chan []byte)
	create := handlers.NewCreateLoop(sdk, chCreate, cfg.Sdk, logger)      //chChreat is chPublish
	mutate := handlers.NewMutator(sdk, chCreate, chMutate, cfg.Sdk, logger)     //chCreate is chSubscribe ; chMutate is chPublish
	transit := handlers.NewTransit(sdk, chMutate, cfg.Sdk, logger)           //chMutate is chSubscribe
	ctx, cancel := context.WithCancel(context.Background())

	// print the contents of cxt
	fmt.Printf("******** CXT content*******: %+v\n", ctx)

	// print the contents of CANCEL
	fmt.Printf("******** CANCEL content*******: %+v\n", cancel)

	// print the contents of cfg
	fmt.Printf("******** CFG content*******: %+v\n", cfg)

	//This code sets up a cancellable context, then runs a series of bootstrap handlers 
	//(responsible for initializing the system's core functionality like SDK, data creation, mutation, etc.). 
	//The use of context ensures that the entire bootstrap process can be canceled gracefully if needed.
	bootstrap.Run(
		ctx,
		cancel,
		cfg,
		[]bootstrap.BootstrapHandler{
			sdk.BootstrapHandler,      //sets up a stream eg: mqtt https://github.com/project-alvarium/alvarium-sdk-go/blob/e5ec0811a099446d00006f5d53f3b054f6733112/pkg/sdk.go#L46
			create.BootstrapHandler,
			mutate.BootstrapHandler,
			transit.BootstrapHandler,
		})
}
