package config

import (
	"fmt"
	loadconfig "github.com/bitxx/load-config"
	"github.com/bitxx/load-config/source"
	"log"
)

type Config struct {
	App       *App    `yaml:"app"`
	Logger    *Logger `yaml:"logger"`
	callbacks []func()
}

func (e *Config) init() {
	e.runCallback()
}

func (e *Config) Init() {
	e.init()
	log.Println("!!! client config init")
}

func (e *Config) runCallback() {
	for i := range e.callbacks {
		e.callbacks[i]()
	}
}

func (e *Config) OnChange() {
	e.init()
	log.Println("!!! client config change and reload")
}

// Setup 载入配置文件
func Setup(s source.Source,
	fs ...func()) {
	_cfg := &Config{
		App:       AppConfig,
		Logger:    LoggerConfig,
		callbacks: fs,
	}
	var err error
	loadconfig.DefaultConfig, err = loadconfig.NewConfig(
		loadconfig.WithSource(s),
		loadconfig.WithEntity(_cfg),
	)
	if err != nil {
		log.Println(fmt.Sprintf("New client config object fail: %s, use default param to start", err.Error()))
		return
	}
	_cfg.Init()
}
