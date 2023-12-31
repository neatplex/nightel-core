package app

import (
	"context"
	"github.com/neatplex/nightel-core/internal/config"
	"github.com/neatplex/nightel-core/internal/database"
	httpServer "github.com/neatplex/nightel-core/internal/http/server"
	"github.com/neatplex/nightel-core/internal/logger"
	"github.com/neatplex/nightel-core/internal/s3"
	"github.com/neatplex/nightel-core/internal/services/container"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

// App integrates the modules to serve.
type App struct {
	context    context.Context
	Config     *config.Config
	Logger     *logger.Logger
	S3         *s3.S3
	HttpServer *httpServer.Server
	Database   *database.Database
	Container  *container.Container
}

// New creates an app from the given configuration file.
func New(configPath string) (a *App, err error) {
	a = &App{}

	a.Config, err = config.New(configPath)
	if err != nil {
		return nil, err
	}

	a.Logger, err = logger.New(a.Config)
	if err != nil {
		return nil, err
	}

	a.Logger.Engine.Debug("config & logger initialized")

	a.Database = database.New(a.Config, a.Logger.Engine)
	a.S3 = s3.New(a.Config, a.Logger.Engine)
	a.Container = container.New(a.Database, a.S3)
	a.HttpServer = httpServer.New(a.Config, a.Logger.Engine, a.Container)

	a.Logger.Engine.Debug("application modules initialized")

	a.setupSignalListener()

	return a, nil
}

// Boot makes sure the critical modules and external sources work fine.
func (a *App) Boot() error {
	a.Database.Connect()
	a.Database.Migrate()
	a.S3.Connect()
	return nil
}

// setupSignalListener sets up a listener to exit signals from os and closes the app gracefully.
func (a *App) setupSignalListener() {
	var cancel context.CancelFunc
	a.context, cancel = context.WithCancel(context.Background())

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-signalChannel
		a.Logger.Engine.Info("system call", zap.String("signal", s.String()))
		cancel()
	}()
}

// Wait avoid dying app and shut it down gracefully on exit signals.
func (a *App) Wait() {
	<-a.context.Done()

	if a.Database != nil {
		a.Database.Close()
	}
	if a.Logger != nil {
		a.Logger.Close()
	}
	if a.HttpServer != nil {
		a.HttpServer.Close()
	}
}
