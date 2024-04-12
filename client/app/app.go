package app

import (
	"encoding/json"
	"errors"
	"ethstats/client/config"
	"ethstats/common/util/cmdutil"
	"ethstats/common/util/connutil"
	"fmt"
	"github.com/bitxx/logger"
	"github.com/bitxx/logger/logbase"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	PingTime    = 5 //second
	PingTimeout = 3 //second
)

type App struct {
	appName     string //名称
	osPlatform  string //平台
	os          string //系统
	version     string //客户端
	readyCh     chan struct{}
	pongCh      chan struct{}
	logger      *logbase.Helper
	delayTicker *time.Timer
	pingTicker  *time.Timer
	procNames   []string
}

func NewApp() *App {
	names := strings.Split(strings.TrimRight(config.AppConfig.ProcNames, ","), ",")
	logInit := logger.NewLogger(
		logger.WithType(config.LoggerConfig.Type),
		logger.WithPath(config.LoggerConfig.Path),
		logger.WithLevel(config.LoggerConfig.Level),
		logger.WithStdout(config.LoggerConfig.Stdout),
		logger.WithCap(config.LoggerConfig.Cap),
	)
	if config.AppConfig.DelayTime <= PingTime {
		logInit.Fatalf("config param 'delayTime' must larger than %d second", PingTime)
	}

	return &App{
		appName:    config.AppConfig.Name,
		osPlatform: runtime.GOARCH,
		os:         runtime.GOOS,
		version:    config.AppConfig.Version,
		procNames:  names,
		readyCh:    make(chan struct{}),
		pongCh:     make(chan struct{}),
		logger:     logInit,
	}
}

func (a *App) Start() {
	// logbase.NewHelper(core.Runtime.GetLogger())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var err error
	isInterrupt := false

	conn := &connutil.ConnWrapper{}
	a.delayTicker = time.NewTimer(0)
	a.pingTicker = time.NewTimer(0)

	defer func() {
		a.close(conn)
		// if not interrupt,restart the client
		if r := recover(); r != nil || !isInterrupt {
			if r != nil {
				a.logger.Warn("conn recover error: ", r)
			}
			time.Sleep(time.Duration(config.AppConfig.DelayTime) * time.Second)
			a.Start()
		}
	}()

	conn, err = connutil.NewDialConn(config.AppConfig.ServerUrl)
	if err != nil {
		a.logger.Warn("dial error: ", err)
		return
	}

	for {
		select {
		case <-a.pingTicker.C:
			if !config.AppConfig.IsPing {
				break
			}
			a.pingTicker.Reset(PingTime * time.Second)
			if err = a.ping(conn); err != nil {
				a.logger.Warn("requested ping failed: ", err)
			}
		case <-a.delayTicker.C:

			//reset the time
			a.delayTicker.Reset(time.Duration(config.AppConfig.DelayTime) * time.Second)

			//request login,need here
			login := map[string][]interface{}{
				"emit": {"hello", map[string]string{
					"id":     a.appName,
					"secret": config.AppConfig.Secret,
				}},
			}
			err := conn.WriteJSON(login)
			if err != nil {
				a.logger.Warn("login request failed: ", err)
				return
			}

			//read info
			go a.readLoop(conn)
		case <-a.readyCh:
			if err = a.reportErrProc(conn); err != nil {
				a.logger.Warn("proc report failed: ", err)
			}
		case <-interrupt:
			a.close(conn)
			isInterrupt = true
			return
		}
	}
}

func (a *App) readLoop(conn *connutil.ConnWrapper) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Warn("readLoop recover error: ", r)
			return
		}
	}()

	//read
	for {
		blob := json.RawMessage{}
		if err := conn.ReadJSON(&blob); err != nil {
			a.logger.Warn("received and decode message error: ", err)
			return
		}
		// Not a system ping, try to decode an actual state message
		var msg map[string][]interface{}
		if err := json.Unmarshal(blob, &msg); err != nil {
			a.logger.Warn("failed to decode message: ", err)
			return
		}

		if len(msg["emit"]) == 0 {
			a.logger.Warn("received message invalid: ", msg)
			return
		}
		msgType, ok := msg["emit"][0].(string)
		if !ok {
			a.logger.Warn("received invalid message type: ", msg["emit"][0])
			return
		}
		a.logger.Trace("received message type: ", msgType)

		switch msgType {
		case "ready":
			//login success
			a.logger.Info("login success!")
			a.readyCh <- struct{}{}
		case "un-authorization":
			//login error
			if len(msg["emit"]) >= 2 {
				if errMsg, ok := msg["emit"][1].(string); ok {
					a.logger.Warn(errMsg)
				}
			}
			return
		case "node-pong":
			//ping pong
			a.pongCh <- struct{}{}
		}

	}
}

func (a *App) ping(conn *connutil.ConnWrapper) error {
	start := time.Now()

	ping := map[string][]interface{}{
		"emit": {"node-ping", map[string]string{
			"id":         config.AppConfig.Name,
			"clientTime": start.String(),
		}},
	}

	if err := conn.WriteJSON(ping); err != nil {
		return err
	}
	a.logger.Trace("send message type: ping")

	// Wait for the pong request to arrive back
	select {
	case <-a.pongCh:
		// Pong delivered, report the latency
	case <-time.After(PingTimeout * time.Second):
		// MsgPing timeout, abort
		return errors.New("ping timed out")
	}

	latency := strconv.Itoa(int((time.Since(start) / time.Duration(1)).Nanoseconds() / 1000000))

	// Send back the measured latency
	a.logger.Trace("sending measured latency: ", latency)

	stats := map[string][]interface{}{
		"emit": {"latency", map[string]string{
			"id":      config.AppConfig.Name,
			"latency": latency,
		}},
	}
	return conn.WriteJSON(stats)
}

// reportErrProc
//
//	@Description: report error proc
//	@receiver a
//	@param conn
//	@return error
func (a *App) reportErrProc(conn *connutil.ConnWrapper) error {
	errProcs := ""
	for _, procName := range a.procNames {
		if procName == "" {
			continue
		}
		_, err := cmdutil.RunCmd(fmt.Sprintf("pidof %s", procName))
		if err != nil {
			a.logger.Error(err)
			errProcs += procName + ","
		}
	}

	// no error
	errProcs = strings.Trim(errProcs, ",")
	if errProcs == "" || len(strings.Split(errProcs, ",")) < 1 {
		return nil
	}

	start := time.Now()
	procReport := map[string][]interface{}{
		"emit": {"proc-report", map[string]string{
			"id":         config.AppConfig.Name,
			"clientTime": start.String(),
			"data":       errProcs,
		}},
	}
	if err := conn.WriteJSON(procReport); err != nil {
		return err
	}
	a.logger.Infof("report error processes names: %s", errProcs)
	return nil
}

func (a *App) close(conn *connutil.ConnWrapper) {
	if conn != nil {
		_ = conn.Close()
	}
	if a.delayTicker != nil {
		_ = a.delayTicker.Stop()
	}
	if a.pingTicker != nil {
		_ = a.pingTicker.Stop()
	}
}
