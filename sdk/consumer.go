package sdk

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/pepper-iot/pulsar-client-go/core/manage"
	"github.com/pepper-iot/pulsar-client-go/core/msg"
	"github.com/rs/zerolog/log"
)

type ConsumerConfig struct {
	Topic string
	Auth  AuthProvider
	manage.SubscriptionMode
	Earliest bool // default is latest, otherwise GETS ALL MESSAGES
}

type consumerImpl struct {
	topic      string
	csm        *manage.ManagedConsumer
	cancelFunc context.CancelFunc
	stopFlag   uint32
	stopped    chan struct{}
}

func (c *consumerImpl) ReceiveAsync(ctx context.Context, queue chan Message) {
	go func() {
		err := c.csm.ReceiveAsync(ctx, queue)
		if err != nil {
			log.Error().Err(err).Str("topic", c.topic).Msg("consumer stopped")
		}
	}()
}

func (c *consumerImpl) ReceiveAndHandle(ctx context.Context, handler PayloadHandler) {
	queue := make(chan Message, 228)
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	go c.ReceiveAsync(ctx, queue)

	for {
		select {
		case <-ctx.Done():
			close(c.stopped)
			return
		case m := <-queue:
			if atomic.LoadUint32(&c.stopFlag) == 1 {
				close(c.stopped)
				return
			}
			log.Debug().Str("topic", c.topic).Msg("consumerImpl receive message")
			bgCtx := context.Background()
			c.Handler(bgCtx, handler, &m)
		}
	}
}

func (c *consumerImpl) Handler(ctx context.Context, handler PayloadHandler, m *Message) {
	diag := map[string]interface{}{}
	diag["msgID"] = m.Msg.GetMessageId()
	diag["topic"] = m.Topic
	defer func(start time.Time) {
		spend := time.Since(start)
		diag["totalSpend"] = spend.String()
		log.Debug().Interface("diag", diag).Msg("Handler trace info")
	}(time.Now())

	now := time.Now()
	var list []*msg.SingleMessage
	var err error
	num := m.Meta.GetNumMessagesInBatch()
	if num > 0 && m.Meta.NumMessagesInBatch != nil {
		list, err = msg.DecodeBatchMessage(m)
		if err != nil {
			log.Error().Err(err).Msg("decode batch message error")
			return
		}
	}
	spend := time.Since(now)
	diag["decodeSpend"] = spend.String()

	now = time.Now()
	if c.csm.Unactive() {
		log.Warn().Str("payload", string(m.Payload)).Str("topic", m.Topic).Msg("unused msg because of consumer is unactivated")
		return
	}
	spend = time.Since(now)
	diag["unactiveSpend"] = spend.String()

	idCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	now = time.Now()
	if c.csm.ConsumerID(idCtx) != m.ConsumerID {
		log.Warn().Str("payload", string(m.Payload)).Str("topic", m.Topic).Msg("unused msg because of consumerID is not match")
		return
	}
	spend = time.Since(now)
	diag["consumerIDSpend"] = spend.String()
	cancel()

	now = time.Now()
	if len(list) == 0 {
		err = handler.HandlePayload(ctx, m, m.Payload)
	} else {
		for i := 0; i < len(list); i++ {
			err = handler.HandlePayload(ctx, m, list[i].SinglePayload)
			if err != nil {
				break
			}
		}
	}
	spend = time.Since(now)
	diag["handlePayloadSpend"] = spend.String()
	if err != nil {
		log.Error().Err(err).Str("payload", string(m.Payload)).Str("topic", m.Topic).Msg("handler failed")
	}

	now = time.Now()
	ackCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	err = c.csm.Ack(ackCtx, *m)
	cancel()
	spend = time.Since(now)
	diag["ackSpend"] = spend.String()
	if err != nil {
		log.Error().Err(err).Str("topic", m.Topic).Msg("ack message failed")
	}

}

func (c *consumerImpl) Stop() {
	atomic.AddUint32(&c.stopFlag, 1)
	c.cancelFunc()
	<-c.stopped
}
