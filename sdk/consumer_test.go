package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestConsumerStop(t *testing.T) {
	accessID := "xxx"
	accessKey := "xxx"
	topic := TopicForAccessID(accessID)

	// create client
	cfg := ClientConfig{
		PulsarAddr: PulsarAddrCN,
	}
	c := NewClient(cfg)

	// create consumer
	csmCfg := ConsumerConfig{
		Topic: topic,
		Auth:  NewAuthProvider(accessID, accessKey),
	}
	csm, _ := c.NewConsumer(csmCfg)

	// handle message
	sleep := time.Second * 5
	h := &helloHandler{
		AesSecret: accessKey[8:24],
		Sleep:     sleep,
	}
	go csm.ReceiveAndHandle(context.Background(), h)

	time.Sleep(5 * time.Second)
	now := time.Now()
	csm.Stop()
	since := time.Since(now)
	log.Debug().Msgf("since: %s", since)
	if since > sleep+time.Second {
		t.Error("stop failed")
	}
}

type helloHandler struct {
	AesSecret string
	Sleep     time.Duration
}

func (h *helloHandler) HandlePayload(ctx context.Context, msg *Message, payload []byte) error {
	log.Info().Interface("messageID", msg.Msg.GetMessageId()).Msg("received message and sleep...")
	time.Sleep(h.Sleep)
	log.Info().Interface("messageID", msg.Msg.GetMessageId()).Msg("finish message")
	return nil
}
