package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nicholaskh/golib/str"
	log "github.com/nicholaskh/log4go"
)

type Cmdline struct {
	Cmd    string
	Params []string
	*Client
}

const (
	CMD_SUBSCRIBE   = "sub"
	CMD_PUBLISH     = "pub"
	CMD_UNSUBSCRIBE = "unsub"
	CMD_HISTORY     = "his"
	CMD_TOKEN       = "gettoken"
	CMD_AUTH_CLIENT = "auth_client"
	CMD_AUTH_SERVER = "auth_server"
	CMD_PING        = "ping"

	OUTPUT_SUBSCRIBED         = "SUBSCRIBED"
	OUTPUT_ALREADY_SUBSCRIBED = "ALREADY SUBSCRIBED"
	OUTPUT_PUBLISHED          = "PUBLISHED"
	OUTPUT_NOT_SUBSCRIBED     = "NOT SUBSCRIBED"
	OUTPUT_UNSUBSCRIBED       = "UNSUBSCRIBED"
	OUTPUT_MESSAGE_PREFIX     = 0
	OUTPUT_COMMAND_PREFIX     = 1
	OUTPUT_DELIMITER          = 2
	OUTPUT_PONG               = "pong"
)

func NewCmdline(input string, cli *Client) (this *Cmdline) {
	this = new(Cmdline)
	parts := strings.SplitN(trimCmdline(input), " ", 3)
	this.Cmd = parts[0]
	this.Params = parts[1:]
	this.Client = cli
	return
}

func (this *Cmdline) Process() (ret string, err error) {
	switch this.Cmd {
	case CMD_SUBSCRIBE:
		if len(this.Params) < 1 || this.Params[0] == "" {
			return "", errors.New("Lack sub channel")
		}
		ret = subscribe(this.Client, this.Params[0])

	case CMD_PUBLISH:
		if len(this.Params) < 2 || this.Params[1] == "" {
			return "", errors.New("Publish without msg\n")
		} else {
			ret = publish(this.Params[0], this.Params[1], false)
		}

	case CMD_UNSUBSCRIBE:
		if len(this.Params) < 1 || this.Params[0] == "" {
			return "", errors.New("Lack unsub channel")
		}
		ret = unsubscribe(this.Client, this.Params[0])

	case CMD_HISTORY:
		if len(this.Params) < 2 {
			return "", errors.New("Invalid Params for history")
		}
		ts, err := strconv.ParseInt(this.Params[1], 10, 64)
		if err != nil {
			return "", err
		}
		channel := this.Params[0]
		hisRet, err := fullHistory(channel, ts)
		if err != nil {
			log.Error(err)
		}

		var retBytes []byte
		retBytes, err = json.Marshal(hisRet)

		ret = string(retBytes)

	//use one appId/secretKey pair
	case CMD_AUTH_SERVER:
		if this.Client.Authed {
			ret = "Already authed"
			err = nil
		} else {
			ret, err = authServer(this.Params[0], this.Params[1])
			if err == nil {
				this.Client.Authed = true
				this.Client.Type = TYPE_SERVER
			}
		}

	case CMD_TOKEN:
		// TODO more secure token generator
		ret = str.Rand(32)
		tokenPool.Set(ret, 1)

	case CMD_AUTH_CLIENT:
		if this.Client.Authed {
			ret = "Already authed"
			err = nil
		} else {
			ret, err = authClient(this.Params[0])
			if err == nil {
				this.Client.Authed = true
				this.Client.Type = TYPE_CLIENT
			}
		}

	case CMD_PING:
		return OUTPUT_PONG, nil

	default:
		return "", errors.New(fmt.Sprintf("Cmd not found: %s\n", this.Cmd))
	}

	return
}

func trimCmdline(str string) string {
	return strings.TrimRight(str, string([]rune{0, 13, 10}))
}
