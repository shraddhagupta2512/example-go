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
	sdk := pkg.NewSdk(annotators, cfg.Sdk, logger)

	// handlers for example functionality
	chCreate := make(chan []byte)
	chMutate := make(chan []byte)
	create := handlers.NewCreateLoop(sdk, chCreate, cfg.Sdk, logger)
	mutate := handlers.NewMutator(sdk, chCreate, chMutate, cfg.Sdk, logger)
	transit := handlers.NewTransit(sdk, chMutate, cfg.Sdk, logger)
	ctx, cancel := context.WithCancel(context.Background())
	bootstrap.Run(
		ctx,
		cancel,
		cfg,
		[]bootstrap.BootstrapHandler{
			sdk.BootstrapHandler,
			create.BootstrapHandler,
			mutate.BootstrapHandler,
			transit.BootstrapHandler,
		})
}
