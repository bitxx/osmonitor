package service

import (
	"encoding/json"
	"ethstats/common/util/connutil"
	"ethstats/common/util/dateutil"
	"ethstats/server/app/model"
	"ethstats/server/config"
	"fmt"
	"github.com/bitxx/logger/logbase"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	messageHello      string = "hello"
	messagePing       string = "node-ping"
	messageProcReport string = "proc-report"
	messageLatency    string = "latency"

	TagErr        = "error info"  //use for tag poolInfo key
	TagProcReport = "proc report" //use for tag poolInfo key
)

// NodeRelay contains the secret used to authenticate the communication between
// the Ethereum node and this server
type NodeRelay struct {
	secret  string
	logger  *logbase.Helper
	channel *model.Channel
}

// NewRelay creates a new NodeRelay struct with required fields
func NewRelay(channel *model.Channel, logger *logbase.Helper) *NodeRelay {
	return &NodeRelay{
		channel: channel,
		secret:  config.ApplicationConfig.Secret,
		logger:  logger,
	}
}

// Close closes the connection between this server and all Ethereum nodes connected to it
func (n *NodeRelay) Close() {
	close(n.channel.MsgPing)
	close(n.channel.MsgLatency)
}

// HandleRequest is the function to handle all server requests that came from
// Ethereum nodes
func (n *NodeRelay) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgradeConn := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := connutil.NewUpgradeConn(upgradeConn, w, r)
	if err != nil {
		n.logger.Warnf("error establishing node connection: %s", err)
		return
	}
	n.logger.Infof("new node connected! (addr=%s, host=%s)", r.RemoteAddr, r.Host)
	go n.loop(conn)
}

// loop loops as long as the connection is alive and retrieves node packages
func (n *NodeRelay) loop(c *connutil.ConnWrapper) {
	errMsg := ""
	// Close connection if an unexpected error occurs and delete the node
	// from the map of connected nodes...
	defer func(c *connutil.ConnWrapper) {
		if n.channel.LoginIDs[c.RemoteAddr().String()] != "" && errMsg != "" {
			n.savePoolInfo(c, TagErr, errMsg)
		}

		//remove error node
		if n.channel.LoginIDs[c.RemoteAddr().String()] != "" {
			delete(n.channel.LoginIDs, c.RemoteAddr().String())
		}

		_ = c.Close()
		n.logger.Warnf("connection with node closed, there are %d connected nodes", len(n.channel.LoginIDs))
	}(c)

	// Client loop
	for {
		_, content, err := c.ReadMessage()
		if err != nil {
			errMsg = fmt.Sprintf("error reading message from client: %s", err)
			return
		}
		// Create emitted message from the node
		msg := model.Message{Content: content}
		msgType, err := msg.GetType()
		if err != nil {
			errMsg = fmt.Sprintf("can't get type of message from the node: %s", err)
			return
		}
		switch msgType {
		case messageHello:
			authMsg, parseError := n.parseAuthMessage(msg)
			if parseError != nil {
				errMsg = fmt.Sprintf("login data parsing error by node[%s], error: %s", authMsg.ID, parseError)
				loginErr := authMsg.SendLoginErrResponse(c, "login data parsing error")
				if loginErr != nil {
					errMsg = fmt.Sprintf("error sending authorization response [parse message error info] to node[%s], error: %s", authMsg.ID, loginErr)
					return
				}
				return
			}
			// first check if the secret is correct
			if authMsg.Secret != n.secret {
				errMsg = fmt.Sprintf("authorization error,invalid secret")
				loginErr := authMsg.SendLoginErrResponse(c, "authorization error,invalid secret")
				if loginErr != nil {
					errMsg = fmt.Sprintf("error sending authorization response [invalid secret] to node[%s], error: %s", authMsg.ID, loginErr)
					return
				}
				return
			}
			//判断节点名称是否重复，遍历效率有点低，有时间了在考虑怎么优化，或者伙计们可以帮忙想个简单的法子
			for k, v := range n.channel.LoginIDs {
				if v == authMsg.ID && k != c.RemoteAddr().String() {
					errMsg = fmt.Sprintf("the id [%s] has login", authMsg.ID)
					n.logger.Errorf("the id [%s] has login", authMsg.ID)
					loginErr := authMsg.SendLoginErrResponse(c, "the login id has being exist,please change the id name")
					if loginErr != nil {
						errMsg = fmt.Sprintf("error sending authorization response [login id is exist] to node[%s], error: %s", authMsg.ID, loginErr)
						return
					}
					return
				}
			}
			sendError := authMsg.SendResponse(c)
			if sendError != nil {
				errMsg = fmt.Sprintf("error sending authorization response to node[%s], error: %s", authMsg.ID, sendError)
				return
			}
			n.channel.LoginIDs[c.RemoteAddr().String()] = authMsg.ID
			n.logger.Infof("node %s login, now %d nodes connected", authMsg.ID, len(n.channel.LoginIDs))
		case messagePing:
			// When the node emit a ping message, we need to respond with pong
			// before five seconds to authorize that node to sent reports
			ping, err := n.parseNodePingMessage(msg)
			if err != nil {
				errMsg = fmt.Sprintf("can't parse ping message sent by node[%s], error: %s", ping.ID, err)
				return
			}
			n.logger.Trace("received message type: ping")
			sendError := ping.SendResponse(c)
			if sendError != nil {
				errMsg = fmt.Sprintf("error sending pong response to node[%s], error: %s", ping.ID, sendError)
				return
			}
			n.logger.Trace("response message type: pong")
			n.channel.MsgPing <- content
		case messageProcReport:
			procReport, err := n.parseProcReportMessage(msg)
			if err != nil {
				errMsg = fmt.Sprintf("get error proc report from node[%s] is wrong, error: %s", procReport.ID, err)
				return
			}
			n.savePoolInfo(c, TagProcReport, "these processes are stopped: "+procReport.Data)
		case messageLatency:
			n.channel.MsgLatency <- content
		}
	}
}

// savePoolInfo
//
//	@Description: save pool info
//	@receiver n
//	@param c
//	@param tag
//	@param content
func (n *NodeRelay) savePoolInfo(c *connutil.ConnWrapper, tag, content string) {
	now := dateutil.ConvertToStr(time.Now(), -1)
	content = "node: [" + n.channel.LoginIDs[c.RemoteAddr().String()] + "-" + c.RemoteAddr().String() + "] " + content
	if len(n.channel.InfoPool[tag]) <= 0 {
		n.channel.InfoPool[tag] = make(map[string]string)
	}
	n.channel.InfoPool[tag][content] = now
}

// parseProcReportMessage
//
//	@Description: proc report
//	@param msg
//	@return *model.ProcReport
//	@return error
func (n *NodeRelay) parseProcReportMessage(msg model.Message) (*model.ProcReport, error) {
	value, err := msg.GetValue()
	if err != nil {
		return &model.ProcReport{}, err
	}
	var report model.ProcReport
	err = json.Unmarshal(value, &report)
	return &report, err
}

// parseNodePingMessage parse the current ping message sent bu the Ethereum node
// and creates a message.NodePing struct with that info
func (n *NodeRelay) parseNodePingMessage(msg model.Message) (*model.NodePing, error) {
	value, err := msg.GetValue()
	if err != nil {
		return &model.NodePing{}, err
	}
	var ping model.NodePing
	err = json.Unmarshal(value, &ping)
	return &ping, err
}

// parseAuthMessage parse the current byte array and transforms it to an AuthMessage struct.
// If an error occurs when json unmarshal, an error is returned
func (n *NodeRelay) parseAuthMessage(msg model.Message) (*model.AuthMessage, error) {
	value, err := msg.GetValue()
	if err != nil {
		return &model.AuthMessage{}, err
	}
	var detail model.AuthMessage
	err = json.Unmarshal(value, &detail)
	return &detail, err
}
