package sdk

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ConsumerList struct {
	list             []*consumerImpl
	FlowPeriodSecond int
	FlowPermit       uint32
	Topic            string
	Stopped          chan struct{}
}

func (l *ConsumerList) ReceiveAndHandle(ctx context.Context, handler PayloadHandler) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(l.list); i++ {
		safe := i
		wg.Add(1)
		go func() {
			l.list[safe].ReceiveAndHandle(ctx, handler)
			wg.Done()
		}()
	}
	go l.CronFlow()
	wg.Wait()
}

func (l *ConsumerList) CronFlow() {
	if l.FlowPeriodSecond == 0 {
		return
	}
	if l.FlowPermit == 0 {
		return
	}
	tk := time.NewTicker(time.Duration(l.FlowPeriodSecond) * time.Second)
	for {
		select {
		case <-l.Stopped:
			tk.Stop()
			log.Info().Str("topic", l.Topic).Msg("consumer cron flow stopped")
			return
		case <-tk.C:
			for i := 0; i < len(l.list); i++ {
				c := l.list[i].csm.Consumer(context.Background())
				if c == nil {
					continue
				}
				if len(c.Overflow) > 0 {
					log.Info().Str("topic", l.Topic).Int("length", len(c.Overflow)).Msg("consumer cron flow overflow")
					_, err := c.RedeliverOverflow(context.Background())
					if err != nil {
						log.Error().Err(err).Str("topic", l.Topic).Msg("consumer cron flow redeliver overflow failed")
					}
				}

				if c.Unactive || len(c.Queue) > 0 {
					continue
				}
				err := c.Flow(l.FlowPermit)
				if err != nil {
					log.Error().Err(err).Str("topic", l.Topic).Msg("consumer cron flow failed")
				}
			}
		}
	}
}

func (l *ConsumerList) Stop() {
	log.Info().Str("topic", l.Topic).Msg("consumer will stop, waiting...")
	close(l.Stopped)
	wg := sync.WaitGroup{}
	for i := 0; i < len(l.list); i++ {
		safei := i
		wg.Add(1)
		go func() {
			l.list[safei].Stop()
			wg.Done()
		}()
	}
	wg.Wait()
	log.Info().Str("topic", l.Topic).Msg("consumer stopped")
}
