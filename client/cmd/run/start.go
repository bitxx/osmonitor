package run

import (
	"ethstats/client/app"
	"ethstats/client/config"
	"ethstats/common/util/textutils"
	"fmt"
	"github.com/bitxx/load-config/source/file"
	"github.com/spf13/cobra"
	"log"
)

var (
	configPath string
	StartCmd   *cobra.Command
)

const (
	name      = "name"
	version   = "version"
	secret    = "secret"
	serverUrl = "server-url"
	delayTime = "delay-time"
	isPing    = "is-ping"
	procNames = "proc-names"
	logPath   = "log-path"
	logLevel  = "log-level"
	logStdout = "log-stdout"
	logType   = "log-type"
	logCap    = "log-cap"
)

func init() {
	StartCmd = &cobra.Command{
		Use:          "start",
		Short:        "run the client",
		Example:      "client start -c settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			config.Setup(
				file.NewSource(file.WithPath(configPath)),
			)
			flag := cmd.PersistentFlags()

			if name, _ := flag.GetString(name); name != "" {
				config.AppConfig.Name = name
			}
			if version, _ := flag.GetString(version); version != "" && config.AppConfig.Version == "" {
				config.AppConfig.Version = version
			}
			if secret, _ := flag.GetString(secret); secret != "" && config.AppConfig.Secret == "" {
				config.AppConfig.Secret = secret
			}
			if serverUrl, _ := flag.GetString(serverUrl); serverUrl != "" && config.AppConfig.ServerUrl == "" {
				config.AppConfig.ServerUrl = serverUrl
			}
			if delayTime, _ := flag.GetUint(delayTime); delayTime > 0 && config.AppConfig.DelayTime <= 0 {
				config.AppConfig.DelayTime = delayTime
			}
			if isPing, _ := flag.GetBool(isPing); isPing && !config.AppConfig.IsPing {
				config.AppConfig.IsPing = isPing
			}
			if procName, _ := flag.GetString(procNames); procName != "" {
				config.AppConfig.ProcNames = procName
			}
			if logPath, _ := flag.GetString(logPath); logPath != "" && config.LoggerConfig.Path == "" {
				config.LoggerConfig.Path = logPath
			}
			if logLevel, _ := flag.GetString(logLevel); logLevel != "" && config.LoggerConfig.Level == "" {
				config.LoggerConfig.Level = logLevel
			}
			if logStdout, _ := flag.GetString(logStdout); logStdout != "" && config.LoggerConfig.Stdout == "" {
				config.LoggerConfig.Stdout = logStdout
			}
			if logType, _ := flag.GetString(logType); logType != "" && config.LoggerConfig.Type == "" {
				config.LoggerConfig.Type = logType
			}
			if logCap, _ := flag.GetUint(logCap); logCap > 0 && config.LoggerConfig.Cap <= 0 {
				config.LoggerConfig.Cap = logCap
			}
			if config.AppConfig.Name == "" {
				log.Fatal("param name can't empty")
			}
			if config.AppConfig.Secret == "" {
				log.Fatal("param secret can't empty")
			}
			if config.AppConfig.ServerUrl == "" {
				log.Fatal("param serverUrl can't empty")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return run()
		},
	}
	cmd := StartCmd.PersistentFlags()
	cmd.StringVarP(&configPath, "config", "c", "", "Start server with provided configuration file")
	cmd.String(name, "", "node name")
	cmd.String(version, "v1.0.0", "version")
	cmd.String(secret, "", "secret")
	cmd.String(serverUrl, "", "server url")
	cmd.Uint(delayTime, 60, "business data transmission interval time")
	cmd.Bool(isPing, false, "turn on/off ping function")
	cmd.String(procNames, "", "monitor proc name, use ',' split multiple proc name")
	cmd.String(logPath, "", "log path")
	cmd.String(logLevel, "trace", "log level")
	cmd.String(logStdout, "default", "default,file")
	cmd.String(logType, "default", "default、zap、logrus")
	cmd.Uint(logCap, 50, "log cap")
}

func run() error {
	logoContent := []byte{10, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 10, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 87, 101, 108, 99, 111, 109, 101, 32, 116, 111, 32, 117, 115, 101, 32, 111, 115, 109, 111, 110, 105, 116, 111, 114, 32, 99, 108, 105, 101, 110, 116, 10, 10, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 95, 111, 111, 79, 111, 111, 95, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 111, 56, 56, 56, 56, 56, 56, 56, 111, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 56, 56, 34, 32, 46, 32, 34, 56, 56, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 40, 124, 32, 45, 95, 45, 32, 124, 41, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 79, 92, 32, 32, 61, 32, 32, 47, 79, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 95, 95, 95, 95, 47, 96, 45, 45, 45, 39, 92, 95, 95, 95, 95, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 46, 39, 32, 32, 92, 92, 124, 32, 32, 32, 32, 32, 124, 47, 47, 32, 32, 96, 46, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 47, 32, 32, 92, 92, 124, 124, 124, 32, 32, 58, 32, 32, 124, 124, 124, 47, 47, 32, 32, 92, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 47, 32, 32, 95, 124, 124, 124, 124, 124, 32, 45, 58, 45, 32, 124, 124, 124, 124, 124, 45, 32, 32, 92, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 124, 32, 32, 32, 124, 32, 92, 92, 92, 32, 32, 45, 32, 32, 47, 47, 47, 32, 124, 32, 32, 32, 124, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 124, 32, 92, 95, 124, 32, 32, 39, 39, 92, 45, 45, 45, 47, 39, 39, 32, 32, 124, 32, 32, 32, 124, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 92, 32, 32, 46, 45, 92, 95, 95, 32, 32, 96, 45, 96, 32, 32, 95, 95, 95, 47, 45, 46, 32, 47, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 95, 95, 95, 96, 46, 32, 46, 39, 32, 32, 47, 45, 45, 46, 45, 45, 92, 32, 32, 96, 46, 32, 46, 32, 95, 95, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 46, 34, 34, 32, 39, 60, 32, 32, 96, 46, 95, 95, 95, 92, 95, 60, 124, 62, 95, 47, 95, 95, 95, 46, 39, 32, 32, 62, 39, 34, 34, 46, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 124, 32, 124, 32, 58, 32, 32, 96, 45, 32, 92, 96, 46, 59, 96, 92, 32, 95, 32, 47, 96, 59, 46, 96, 47, 32, 45, 32, 96, 32, 58, 32, 124, 32, 124, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 92, 32, 32, 92, 32, 96, 45, 46, 32, 32, 32, 92, 95, 32, 95, 95, 92, 32, 47, 95, 95, 32, 95, 47, 32, 32, 32, 46, 45, 96, 32, 47, 32, 32, 47, 10, 32, 32, 32, 32, 61, 61, 61, 61, 61, 61, 96, 45, 46, 95, 95, 95, 95, 96, 45, 46, 95, 95, 95, 92, 95, 95, 95, 95, 95, 47, 95, 95, 95, 46, 45, 96, 95, 95, 95, 95, 46, 45, 39, 61, 61, 61, 61, 61, 61, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 96, 61, 45, 45, 45, 61, 39, 10, 32, 32, 32, 32, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 66, 117, 100, 100, 104, 97, 32, 98, 108, 101, 115, 115, 32, 32, 32, 32, 32, 32, 66, 117, 103, 32, 98, 108, 101, 115, 115, 10, 10, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 10}
	fmt.Println(textutils.Red(string(logoContent)))

	app.NewApp().Start()
	return nil
}
