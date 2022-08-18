// Copyright 2018 Comcast Cable Communications Management, LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conn

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pepper-iot/pulsar-client-go/core/frame"
	"github.com/pepper-iot/pulsar-client-go/pkg/api"
	"github.com/pepper-iot/pulsar-client-go/utils"
)

// NewConnector returns a ready-to-use connector.
func NewConnector(s frame.CmdSender, dispatcher *frame.Dispatcher, ac AuthConfig) *Connector {
	return &Connector{
		S:          s,
		Dispatcher: dispatcher,
		AuthConfig: ac,
	}
}

type AuthConfig struct {
	AuthMethod string
	AuthData   []byte
}

// connector encapsulates the logic for the CONNECT <-> (CONNECTED|ERROR)
// request-response cycle.
//
// https://pulsar.incubator.apache.org/docs/latest/project/BinaryProtocol/#Connectionestablishment-ly8l2n
type Connector struct {
	S          frame.CmdSender
	Dispatcher *frame.Dispatcher // used to manage the request/response state
	AuthConfig AuthConfig
}

// Connect initiates the client's session. After sending,
// the client should wait for a `Connected` or `Error`
// response from the server.
//
// The provided context should have a timeout associated with it.
//
// It's required to have completed Connect/Connected before using the client.
func (c *Connector) Connect(ctx context.Context, authMethod, proxyBrokerURL string) (*api.CommandConnected, error) {
	resp, cancel, err := c.Dispatcher.RegisterGlobal()
	if err != nil {
		return nil, err
	}
	defer cancel()

	// NOTE: The source seems to indicate that the ERROR messages's
	// RequestID will be -1 (ie UndefRequestID) in the case that it's
	// associated with a CONNECT request.
	// https://github.com/apache/incubator-pulsar/blob/fdc7b8426d8253c9437777ae51a4639239550f00/pulsar-broker/src/main/java/org/apache/pulsar/broker/service/ServerCnx.java#L325
	errResp, cancel, err := c.Dispatcher.RegisterReqID(utils.UndefRequestID)
	if err != nil {
		return nil, err
	}
	defer cancel()

	// create and send CONNECT msg

	connect := api.CommandConnect{
		ClientVersion:   proto.String(utils.ClientVersion),
		ProtocolVersion: proto.Int32(utils.ProtoVersion),
	}
	if authMethod != "" {
		connect.AuthMethodName = proto.String(authMethod)
	}
	if proxyBrokerURL != "" {
		addr := strings.TrimPrefix(proxyBrokerURL, "pulsar://")
		connect.ProxyToBrokerUrl = proto.String(addr)
	}

	if c.AuthConfig.AuthMethod != "" {
		connect.AuthMethodName = proto.String(c.AuthConfig.AuthMethod)
	}

	if c.AuthConfig.AuthData != nil {
		connect.AuthData = c.AuthConfig.AuthData
	}

	cmd := api.BaseCommand{
		Type:    api.BaseCommand_CONNECT.Enum(),
		Connect: &connect,
	}

	if err := c.S.SendSimpleCmd(cmd); err != nil {
		return nil, err
	}

	// wait for the response, error, or timeout

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case connectedFrame := <-resp:
		return connectedFrame.BaseCmd.GetConnected(), nil

	case errFrame := <-errResp:
		err := errFrame.BaseCmd.GetError()
		return nil, fmt.Errorf("%s: %s", err.GetError().String(), err.GetMessage())
	}
}
