package ingest

import (
	"context"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Consumer struct {
	client mqtt.Client
	pool   *WorkerPool
	logger *slog.Logger
	topic  string
}

func NewConsumer(brokerURL, clientID string, pool *WorkerPool, logger *slog.Logger) *Consumer {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetCleanSession(false) // persistent session
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	opts.SetKeepAlive(30 * time.Second)

	sharedTopic := "$share/fp/nova/+/+/+/#"

	c := &Consumer{
		pool:   pool,
		logger: logger,
		topic:  sharedTopic,
	}

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Info("Connected to MQTT broker; subscribing to shared topic", "topic", sharedTopic)
		token := client.Subscribe(sharedTopic, 1, func(cl mqtt.Client, msg mqtt.Message) {
			payloadCopy := make([]byte, len(msg.Payload()))
			copy(payloadCopy, msg.Payload())

			if !pool.Submit(payloadCopy) {
				logger.Warn("Failed to submit event to worker pool: queue full")
			}
		})
		if token.Wait() && token.Error() != nil {
			logger.Error("Failed to subscribe to MQTT shared topic", "topic", sharedTopic, "error", token.Error())
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Warn("MQTT connection lost, reconnecting...", "error", err)
	})

	c.client = mqtt.NewClient(opts)
	return c
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Starting MQTT ingestion consumer")
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		c.logger.Warn("Initial MQTT connection attempt deferred/failed (will auto-retry)", "error", token.Error())
	}

	<-ctx.Done()
	c.logger.Info("Disconnecting MQTT consumer...")
	c.client.Disconnect(250)
	return nil
}
