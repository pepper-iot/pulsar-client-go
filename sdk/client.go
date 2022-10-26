package sdk

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/pepper-iot/pulsar-client-go/core/manage"
	"github.com/pepper-iot/pulsar-client-go/core/msg"
)

const (
	DefaultFlowPeriodSecond = 30
	DefaultFlowPermit       = 10

	PulsarAddrCN = "pulsar://mqe.tuyacn.com:7285"
	PulsarAddrEU = "pulsar://mqe.tuyaeu.com:7285"
	PulsarAddrUS = "pulsar://mqe.tuyaus.com:7285"
)

type Message = msg.Message

type Client interface {
	NewConsumer(config ConsumerConfig) (Consumer, error)
}

type Consumer interface {
	ReceiveAndHandle(ctx context.Context, handler PayloadHandler)
	Stop()
}

type PayloadHandler interface {
	HandlePayload(ctx context.Context, msg *Message, payload []byte) error
}

type ClientConfig struct {
	PulsarAddr string
}

type client struct {
	pool *manage.ClientPool
	Addr string
}

func NewClient(cfg ClientConfig) Client {
	return &client{
		pool: manage.NewClientPool(),
		Addr: cfg.PulsarAddr,
	}
}

func subscriptionName(topic string) string {
	envName := os.Getenv("SUBSCRIPTION_NAME")
	if envName != "" {
		return envName
	}
	return getTenant(topic) + "-sub"
}

func getTenant(topic string) string {
	topic = strings.TrimPrefix(topic, "persistent://")
	end := strings.Index(topic, "/")
	return topic[:end]
}

func (c *client) NewConsumer(config ConsumerConfig) (Consumer, error) {
	log.Info().Str("pulsar", c.Addr).Interface("config", config).Msg("start creating consumer")

	errs := make(chan error, 10)
	go func() {
		for err := range errs {
			log.Error().Err(err).Msg("error creating consumer")
		}
	}()
	cfg := manage.ConsumerConfig{
		ClientConfig: manage.ClientConfig{
			Addr:       c.Addr,
			AuthData:   config.Auth.AuthData(),
			AuthMethod: config.Auth.AuthMethod(),
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Errs: errs,
		},
		Topic:              config.Topic,
		SubMode:            config.SubscriptionMode,
		Earliest:           config.Earliest,
		Name:               subscriptionName(config.Topic),
		NewConsumerTimeout: time.Minute,
	}
	p := c.GetPartition(config.Topic, cfg.ClientConfig)

	// partitioned topic
	if p > 0 {
		list := make([]*consumerImpl, 0, p)
		originTopic := cfg.Topic
		for i := 0; i < p; i++ {
			cfg.Topic = fmt.Sprintf("%s-partition-%d", originTopic, i)
			mc := manage.NewManagedConsumer(c.pool, cfg)
			list = append(list, &consumerImpl{
				csm:     mc,
				topic:   cfg.Topic,
				stopped: make(chan struct{}),
			})
		}
		consumerList := &ConsumerList{
			list:             list,
			FlowPeriodSecond: DefaultFlowPeriodSecond,
			FlowPermit:       DefaultFlowPermit,
			Topic:            config.Topic,
			Stopped:          make(chan struct{}),
		}
		return consumerList, nil
	}

	// single topic
	mc := manage.NewManagedConsumer(c.pool, cfg)
	log.Info().Str("pulsar", c.Addr).Interface("config", config).Msg("created consumer")
	return &consumerImpl{
		csm:     mc,
		topic:   cfg.Topic,
		stopped: make(chan struct{}),
	}, nil

}

func (c *client) GetPartition(topic string, config manage.ClientConfig) int {
	p, err := c.pool.Partitions(context.Background(), config, topic)
	if err != nil {
		return 0
	}
	return int(p.GetPartitions())
}

func TopicForAccessID(accessID string) string {
	topic := fmt.Sprintf("persistent://%s/out/event", accessID)
	return topic
}
