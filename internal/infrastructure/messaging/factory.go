package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"skoolz/config"
	"skoolz/internal/logger"
)

// MessagingFactory manages Kafka and NATS clients
type MessagingFactory struct {
	config      *config.Config
	logger      *slog.Logger
	startupLog  *logger.StartupLogger
	kafkaClient *KafkaClient
	// natsClient  *NatsClient
	mu sync.RWMutex
}

// NewMessagingFactory creates a new messaging factory
func NewMessagingFactory(cfg *config.Config, log *slog.Logger) *MessagingFactory {
	return &MessagingFactory{
		config:     cfg,
		logger:     log,
		startupLog: logger.NewStartupLogger(log),
	}
}

// GetKafkaClient returns or creates a Kafka client
func (m *MessagingFactory) GetKafkaClient() (*KafkaClient, error) {
	// Check if Kafka is enabled
	if !m.config.Kafka.EnableKafka {
		return nil, fmt.Errorf("kafka is disabled in configuration")
	}

	m.mu.RLock()
	if m.kafkaClient != nil {
		defer m.mu.RUnlock()
		return m.kafkaClient, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if m.kafkaClient != nil {
		return m.kafkaClient, nil
	}

	client, err := NewKafkaClient(&m.config.Kafka, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	m.kafkaClient = client
	return client, nil
}

// GetNatsClient returns or creates a NATS client
// func (m *MessagingFactory) GetNatsClient() (*NatsClient, error) {
// 	m.mu.RLock()
// 	if m.natsClient != nil {
// 		defer m.mu.RUnlock()
// 		return m.natsClient, nil
// 	}
// 	m.mu.RUnlock()

// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	// Double-check after acquiring write lock
// 	if m.natsClient != nil {
// 		return m.natsClient, nil
// 	}

// 	client, err := NewNatsClient(&m.config.Nats, m.logger)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create NATS client: %w", err)
// 	}

// 	m.natsClient = client
// 	return client, nil
// }

// InitializeStreams initializes JetStream streams for NATS
// func (m *MessagingFactory) InitializeStreams() error {
// 	natsClient, err := m.GetNatsClient()
// 	if err != nil {
// 		return fmt.Errorf("failed to get NATS client for stream initialization: %w", err)
// 	}

// 	// Create streams for different message types
// 	streams := map[string][]string{
// 		"USER_EVENTS": {
// 			m.config.Nats.Subjects.UserEvents,
// 			m.config.Nats.Subjects.UserCommands,
// 		},
// 		"USER_QUERIES": {
// 			m.config.Nats.Subjects.UserQueries,
// 		},
// 	}

// 	// Retry mechanism for stream creation
// 	maxRetries := 3
// 	for i := 0; i < maxRetries; i++ {
// 		success := true
// 		for streamName, subjects := range streams {
// 			if err := natsClient.CreateStream(streamName, subjects); err != nil {
// 				m.startupLog.Warning("Stream creation failed, retrying", "stream", streamName, "attempt", i+1, "error", err.Error())
// 				success = false
// 				break
// 			}
// 			m.startupLog.Stream("Stream ready", "name", streamName)
// 		}

// 		if success {
// 			break
// 		}

// 		if i < maxRetries-1 {
// 			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
// 		} else {
// 			return fmt.Errorf("failed to create streams after %d attempts", maxRetries)
// 		}
// 	}

// 	// Add a small delay to ensure streams are fully ready
// 	time.Sleep(100 * time.Millisecond)

// 	return nil
// }

// StartSubscriptions starts all message subscriptions
func (m *MessagingFactory) StartSubscriptions(ctx context.Context) error {
	// Start Kafka subscriptions
	if err := m.startKafkaSubscriptions(ctx); err != nil {
		return fmt.Errorf("failed to start Kafka subscriptions: %w", err)
	}

	// Start NATS subscriptions
	// if err := m.startNatsSubscriptions(ctx); err != nil {
	// 	return fmt.Errorf("failed to start NATS subscriptions: %w", err)
	// }

	return nil
}

// startKafkaSubscriptions starts Kafka message subscriptions
func (m *MessagingFactory) startKafkaSubscriptions(ctx context.Context) error {
	// Check if Kafka is enabled
	if !m.config.Kafka.EnableKafka {
		m.startupLog.Info("Kafka disabled, skipping subscriptions")
		return nil
	}

	kafkaClient, err := m.GetKafkaClient()
	if err != nil {
		return fmt.Errorf("failed to get Kafka client: %w", err)
	}

	// Start subscription for user events
	go func() {
		if err := kafkaClient.Subscribe(ctx, m.config.Kafka.Topics.UserEvents, func(msg KafkaMessage) error {
			return m.handleKafkaUserEvent(msg)
		}); err != nil {
			m.logger.Error("Kafka subscription error", "error", err.Error(), "topic", m.config.Kafka.Topics.UserEvents)
		}
	}()

	// Start subscription for user commands
	go func() {
		if err := kafkaClient.Subscribe(ctx, m.config.Kafka.Topics.UserCommands, func(msg KafkaMessage) error {
			return m.handleKafkaUserCommand(msg)
		}); err != nil {
			m.logger.Error("Kafka subscription error", "error", err.Error(), "topic", m.config.Kafka.Topics.UserCommands)
		}
	}()

	return nil
}

// startNatsSubscriptions starts NATS message subscriptions
// func (m *MessagingFactory) startNatsSubscriptions(ctx context.Context) error {
// 	natsClient, err := m.GetNatsClient()
// 	if err != nil {
// 		return fmt.Errorf("failed to get NATS client: %w", err)
// 	}

// 	// Start subscription for user events
// 	go func() {
// 		m.startupLog.Subscription("Subscribing to user events", "subject", m.config.Nats.Subjects.UserEvents)
// 		if err := natsClient.Subscribe(ctx, m.config.Nats.Subjects.UserEvents, m.config.Nats.QueueGroup, func(msg NatsMessage) error {
// 			return m.handleNatsUserEvent(msg)
// 		}); err != nil {
// 			m.startupLog.Error("NATS subscription failed", "error", err.Error(), "subject", m.config.Nats.Subjects.UserEvents)
// 		}
// 	}()

// 	// Start subscription for user commands
// 	go func() {
// 		m.startupLog.Subscription("Subscribing to user commands", "subject", m.config.Nats.Subjects.UserCommands)
// 		if err := natsClient.Subscribe(ctx, m.config.Nats.Subjects.UserCommands, m.config.Nats.QueueGroup, func(msg NatsMessage) error {
// 			return m.handleNatsUserCommand(msg)
// 		}); err != nil {
// 			m.startupLog.Error("NATS subscription failed", "error", err.Error(), "subject", m.config.Nats.Subjects.UserCommands)
// 		}
// 	}()

// 	// Start subscription for user queries
// 	go func() {
// 		m.startupLog.Subscription("Subscribing to user queries", "subject", m.config.Nats.Subjects.UserQueries)
// 		if err := natsClient.Subscribe(ctx, m.config.Nats.Subjects.UserQueries, m.config.Nats.QueueGroup, func(msg NatsMessage) error {
// 			return m.handleNatsUserQuery(msg)
// 		}); err != nil {
// 			m.startupLog.Error("NATS subscription failed", "error", err.Error(), "subject", m.config.Nats.Subjects.UserQueries)
// 		}
// 	}()

// 	return nil
// }

// Message handlers
func (m *MessagingFactory) handleKafkaUserEvent(msg KafkaMessage) error {
	m.logger.Info("Processing Kafka user event", "topic", msg.Topic, "key", msg.Key, "value", msg.Value)
	// TODO: Implement user event processing logic
	return nil
}

func (m *MessagingFactory) handleKafkaUserCommand(msg KafkaMessage) error {
	m.logger.Info("Processing Kafka user command", "topic", msg.Topic, "key", msg.Key, "value", msg.Value)
	// TODO: Implement user command processing logic
	return nil
}

// func (m *MessagingFactory) handleNatsUserEvent(msg NatsMessage) error {
// 	m.logger.Info("Processing NATS user event", "subject", msg.Subject, "data", msg.Data)
// 	// TODO: Implement user event processing logic
// 	return nil
// }

// func (m *MessagingFactory) handleNatsUserCommand(msg NatsMessage) error {
// 	m.logger.Info("Processing NATS user command", "subject", msg.Subject, "data", msg.Data)
// 	// TODO: Implement user command processing logic
// 	return nil
// }

// func (m *MessagingFactory) handleNatsUserQuery(msg NatsMessage) error {
// 	m.logger.Info("Processing NATS user query", "subject", msg.Subject, "data", msg.Data)
// 	// TODO: Implement user query processing logic
// 	return nil
// }

// Close closes all messaging clients
func (m *MessagingFactory) Close() error {
	var errs []error

	if m.kafkaClient != nil {
		if err := m.kafkaClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka client: %w", err))
		}
	}

	// if m.natsClient != nil {
	// 	if err := m.natsClient.Close(); err != nil {
	// 		errs = append(errs, fmt.Errorf("failed to close NATS client: %w", err))
	// 	}
	// }

	if len(errs) > 0 {
		return fmt.Errorf("errors closing messaging factory: %v", errs)
	}

	return nil
}

// HealthCheck checks the health of all messaging clients
func (m *MessagingFactory) HealthCheck() error {
	var errs []error

	if m.kafkaClient != nil {
		if err := m.kafkaClient.HealthCheck(); err != nil {
			errs = append(errs, fmt.Errorf("kafka health check failed: %w", err))
		}
	}

	// if m.natsClient != nil {
	// 	if err := m.natsClient.HealthCheck(); err != nil {
	// 		errs = append(errs, fmt.Errorf("NATS health check failed: %w", err))
	// 	}
	// }

	if len(errs) > 0 {
		return fmt.Errorf("messaging health check failed: %v", errs)
	}

	return nil
}
