package service

import (
	"ethstats/common/util/connutil"
	"ethstats/common/util/emailutil"
	"ethstats/server/app/model"
	"ethstats/server/config"
	"fmt"
	"github.com/bitxx/logger/logbase"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

// Api is the responsible to send node state to registered hub
type Api struct {
	logger *logbase.Helper
	hub    *hub
}

// NewApi creates a new Api struct with the required service
func NewApi(channel *model.Channel, logger *logbase.Helper) *Api {
	hub := &hub{
		register: make(chan *connutil.ConnWrapper),
		logger:   logger,
		close:    make(chan interface{}),
		clients:  make(map[*connutil.ConnWrapper]bool),
		channel:  channel,
	}
	go hub.loop()
	return &Api{
		logger: logger,
		hub:    hub,
	}
}

// Close this server and all registered client connections
func (a *Api) Close() {
	a.logger.Info("prepared to close all client connections")
	a.hub.close <- "close"
}

// HandleRequest handle all request from hub that are not Ethereum nodes
func (a *Api) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgradeConn := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := connutil.NewUpgradeConn(upgradeConn, w, r)
	if err != nil {
		a.logger.Errorf("error trying to establish communication with client (addr=%s, host=%s, URI=%s), %s",
			r.RemoteAddr, r.Host, r.RequestURI, err)
		return
	}
	a.logger.Infof("connected new client! (host=%s)", r.Host)
	a.hub.register <- conn
}

// hub maintain a list of registered clients to send messages
type hub struct {
	register chan *connutil.ConnWrapper
	logger   *logbase.Helper
	close    chan interface{}
	clients  map[*connutil.ConnWrapper]bool
	channel  *model.Channel
}

// loop loops as the server is alive and send messages to registered clients
func (h *hub) loop() {
	poolInfoTicker := time.NewTicker(time.Duration(config.EmailConfig.DelayTime) * time.Second)
	defer func() {
		poolInfoTicker.Stop()
	}()
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case ping := <-h.channel.MsgPing:
			//debug log for show the ping
			//h.logger.Info("debug log show ping = > ", string(ping))
			//use for send to any fronted client
			h.writeMessage(ping)
		case latency := <-h.channel.MsgLatency:
			//debug log for show the latency
			//h.logger.Info("debug log show latency = > ", string(latency))
			//use for send to any fronted client
			h.writeMessage(latency)
		case <-poolInfoTicker.C:
			if len(h.channel.InfoPool) <= 0 {
				break
			}
			msg := ""
			for tag, infos := range h.channel.InfoPool {
				msg += tag + ":\n"
				for info, latestTime := range infos {
					msg += latestTime + " => " + info + "\n"
				}
				msg += "\n"
			}

			h.channel.InfoPool = make(map[string]map[string]string) //clean cache

			err := emailutil.SendEmailDefault(fmt.Sprintf("%s-monitor report\n", time.Now().Format("2006-01-02 15:04:05")), msg)
			if err != nil {
				h.logger.Errorf("send email error: %s, email info: \n%s", err, msg)
			}
		case <-h.close:
			h.quit()
			break
		}
	}
}

// writeMessage to all registered clients. If an error occurs sending a message to a client,
// then these connection is closed and removed from the pool of registered clients
func (h *hub) writeMessage(msg []byte) {
	for client := range h.clients {
		err := client.WriteMessage(1, msg)
		if err != nil {
			h.logger.Infof("Closed connection with client: %s", client.RemoteAddr())
			// close and delete the client connection and release
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *hub) quit() {
	h.logger.Info("Closing all registered clients")
	for client := range h.clients {
		client.Close()
		delete(h.clients, client)
	}
	close(h.register)
	close(h.close)
}
