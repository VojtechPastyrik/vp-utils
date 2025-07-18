package load

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/rabbitmq"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	rabbitmqUtisl "github.com/VojtechPastyrik/vp-utils/utils/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var (
	FlagHost              string
	FlagPort              int
	FlagUser              string
	FlagPassword          string
	FlagVirtualHost       string
	FlagSsl               bool
	FlagSslCert           string
	FlagSslKey            string
	FlagDuration          string
	FlagQueueCount        int
	FlagExchangeCount     int
	FlagRoutingKeyCount   int
	FlagMessageSize       int
	FlagParallelClients   int
	sentMessagesCount     int32
	receivedMessagesCount int32
)

func init() {
	logger.Initialize(logger.InfoLevel)
}

var Cmd = &cobra.Command{
	Use:     "load-test",
	Aliases: []string{"lt"},
	Short:   "Run RabbitMQ load test",
	Long:    "Run RabbitMQ load test using the specified configuration. This command is useful for performance testing and benchmarking RabbitMQ servers.",
	Run: func(cmd *cobra.Command, args []string) {
		runLoadTest(FlagHost, FlagPort, FlagUser, FlagPassword, FlagVirtualHost, FlagSsl, FlagSslCert, FlagSslKey, FlagDuration, FlagQueueCount, FlagExchangeCount, FlagRoutingKeyCount, FlagMessageSize, FlagParallelClients)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagHost, "host", "H", "localhost", "RabbitMQ host")
	Cmd.MarkFlagRequired("host")
	Cmd.Flags().IntVarP(&FlagPort, "port", "P", 5672, "RabbitMQ port")
	Cmd.MarkFlagRequired("port")
	Cmd.Flags().StringVarP(&FlagUser, "user", "u", "guest", "RabbitMQ username")
	Cmd.MarkFlagRequired("user")
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "guest", "RabbitMQ password")
	Cmd.MarkFlagRequired("password")
	Cmd.Flags().StringVarP(&FlagVirtualHost, "vhost", "v", "/", "RabbitMQ virtual host")
	Cmd.Flags().BoolVarP(&FlagSsl, "ssl", "s", false, "Enable SSL")
	Cmd.Flags().StringVarP(&FlagSslCert, "ssl-cert", "c", "", "Path to SSL certificate")
	Cmd.Flags().StringVarP(&FlagSslKey, "ssl-key", "k", "", "Path to SSL key")
	Cmd.Flags().StringVarP(&FlagDuration, "duration", "d", "1m", "Duration of the load test")
	Cmd.Flags().IntVarP(&FlagQueueCount, "queue-count", "q", 50, "Number of queues to create")
	Cmd.Flags().IntVarP(&FlagExchangeCount, "exchange-count", "e", 20, "Number of exchanges to create")
	Cmd.Flags().IntVarP(&FlagRoutingKeyCount, "routing-keys", "r", 50, "Number of routing keys to use")
	Cmd.Flags().IntVarP(&FlagMessageSize, "message-size", "m", 1024, "Size of each message in bytes")
	Cmd.Flags().IntVarP(&FlagParallelClients, "parallel-clients", "C", 20, "Number of parallel clients")
}

func runLoadTest(host string, port int, user, password, virtualHost string, ssl bool, sslCert, sslKey, duration string, queueCount, exchangeCount, routingKeyCount, messageSize, parallelClients int) {
	con, ch, err := rabbitmqUtisl.ConnectToRabbitMQ(ssl, user, password, host, port, virtualHost, sslCert, sslKey)
	if err != nil {
		logger.Errorf("connection to rabbitmq failed: %v", err)
		return
	}

	exchangeList, queueList, exchangeRoutingKeys := createResources(ch, exchangeCount, queueCount, routingKeyCount)
	defer func() {
		cleanupResources(ch, exchangeList, queueList)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("signal received, cleaning up resources...")
		cleanupResources(ch, exchangeList, queueList)
		os.Exit(0)
	}()

	durationParsed, err := time.ParseDuration(duration)
	if err != nil {
		logger.Errorf("invalid duration format: %v", err)
		return
	}

	startConsumersAndProducers(con, queueList, exchangeList, exchangeRoutingKeys, parallelClients, durationParsed, messageSize)

	logger.Infof("load test completed. sent messages: %d, received messages: %d", sentMessagesCount, receivedMessagesCount)
}

func createResources(ch *amqp091.Channel, exchangeCount, queueCount, routingKeyCount int) ([]string, []string, map[string][]string) {
	rand.Seed(time.Now().UnixNano())

	exchangeList := make([]string, exchangeCount)
	queueList := make([]string, queueCount)
	routingKeys := make([]string, routingKeyCount)
	exchangeRoutingKeys := make(map[string][]string)

	for i := 0; i < exchangeCount; i++ {
		exchangeName := fmt.Sprintf("test-exchange-%d", i)
		err := ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil)
		if err != nil {
			logger.Errorf("failed to declare exchange %s: %v", exchangeName, err)
		}
		exchangeList[i] = exchangeName
	}

	for i := 0; i < queueCount; i++ {
		queueName := fmt.Sprintf("test-queue-%d", i)
		_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			logger.Errorf("failed to declare queue %s: %v", queueName, err)
		}
		queueList[i] = queueName
	}

	for i := 0; i < routingKeyCount; i++ {
		routingKeys[i] = fmt.Sprintf("test-routing-key-%d", i)
	}

	// Bind each queue to each exchange with all routing keys for maximum message duplication
	for _, exchange := range exchangeList {
		for _, routingKey := range routingKeys {
			for _, queue := range queueList {
				err := ch.QueueBind(queue, routingKey, exchange, false, nil)
				if err != nil {
					logger.Errorf("failed to bind queue %s to exchange %s with routing key %s: %v", queue, exchange, routingKey, err)
				}
			}
			if !contains(exchangeRoutingKeys[exchange], routingKey) {
				exchangeRoutingKeys[exchange] = append(exchangeRoutingKeys[exchange], routingKey)
			}
		}
	}

	return exchangeList, queueList, exchangeRoutingKeys
}

func cleanupResources(ch *amqp091.Channel, exchangeList, queueList []string) {
	for _, exchange := range exchangeList {
		err := ch.ExchangeDelete(exchange, false, false)
		if err != nil {
			logger.Errorf("failed to delete exchange %s: %v", exchange, err)
		}
	}

	for _, queue := range queueList {
		_, err := ch.QueueDelete(queue, false, false, false)
		if err != nil {
			logger.Errorf("failed to delete queue %s: %v", queue, err)
		}
	}
}

func startConsumersAndProducers(conn *amqp091.Connection, queueList, exchangeList []string, exchangeRoutingKeys map[string][]string, parallelClients int, duration time.Duration, messageSize int) {
	var wg sync.WaitGroup
	doneChan := make(chan struct{})

	logger.Info("starting consumers...")
	assignedQueues := assignQueuesToConsumers(queueList, parallelClients)
	for i := 0; i < parallelClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Infof("consumer %d started", id)
			ch, err := conn.Channel()
			if err != nil {
				logger.Errorf("failed to create channel for consumer %d: %v", id, err)
				return
			}
			defer ch.Close()

			logger.Infof("consumer %d assigned queues: %v", id, assignedQueues[id])
			consumeMessages(ch, assignedQueues[id], duration, doneChan)
			logger.Infof("consumer %d finished", id)
		}(i)
	}

	logger.Info("starting producers...")
	// Assign exchanges to producers to ensure all exchanges are used
	assignedExchanges := assignExchangesToProducers(exchangeList, parallelClients)
	for i := 0; i < parallelClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Infof("producer %d started", id)
			ch, err := conn.Channel()
			if err != nil {
				logger.Errorf("failed to create channel for producer %d: %v", id, err)
				return
			}
			defer ch.Close()

			producerExchanges := assignedExchanges[id]
			logger.Infof("producer %d assigned exchanges: %v", id, producerExchanges)
			produceMessages(ch, producerExchanges, exchangeRoutingKeys, messageSize, duration, doneChan)
			logger.Infof("producer %d finished", id)
		}(i)
	}

	go func() {
		time.Sleep(duration)
		logger.Info("duration expired, stopping load test...")
		close(doneChan)
	}()

	logger.Info("waiting for all goroutines to finish...")
	wg.Wait()
	logger.Info("all producers and consumers stopped")
}

func assignQueuesToConsumers(queueList []string, parallelClients int) [][]string {
	assigned := make([][]string, parallelClients)
	for i, queue := range queueList {
		assigned[i%parallelClients] = append(assigned[i%parallelClients], queue)
	}
	return assigned
}

func assignExchangesToProducers(exchangeList []string, parallelClients int) [][]string {
	assigned := make([][]string, parallelClients)
	for i, exchange := range exchangeList {
		assigned[i%parallelClients] = append(assigned[i%parallelClients], exchange)
	}
	return assigned
}

func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func consumeMessages(ch *amqp091.Channel, queues []string, duration time.Duration, doneChan <-chan struct{}) {
	var wg sync.WaitGroup
	timeout := time.After(duration)

	for _, queue := range queues {
		wg.Add(1)
		go func(queue string) {
			defer wg.Done()

			// Set QoS to prefetch messages for better performance
			err := ch.Qos(100, 0, false)
			if err != nil {
				logger.Errorf("failed to set QoS for queue %s: %v", queue, err)
				return
			}

			// Use Consume instead of Get for better throughput
			msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
			if err != nil {
				logger.Errorf("failed to consume from queue %s: %v", queue, err)
				return
			}

			for {
				select {
				case <-doneChan:
					logger.Infof("consumer received stop signal for queue %s, exiting...", queue)
					drainQueues(ch, []string{queue})
					return
				case <-timeout:
					logger.Infof("consumer timeout reached for queue %s, exiting...", queue)
					drainQueues(ch, []string{queue})
					return
				case delivery, ok := <-msgs:
					if !ok {
						logger.Infof("message channel closed for queue %s", queue)
						return
					}
					atomic.AddInt32(&receivedMessagesCount, 1)
					_ = delivery // Use delivery to avoid unused variable warning
				}
			}
		}(queue)
	}

	wg.Wait()
}

func drainQueues(ch *amqp091.Channel, queues []string) {
	for _, queue := range queues {
		// Set high QoS for fast draining
		err := ch.Qos(500, 0, false)
		if err != nil {
			logger.Errorf("failed to set QoS for draining queue %s: %v", queue, err)
			continue
		}

		// Use Consume for faster draining
		msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
		if err != nil {
			logger.Errorf("failed to consume from queue %s for draining: %v", queue, err)
			continue
		}

		// Drain queue until empty with idle timeout
		idleTimer := time.NewTimer(500 * time.Millisecond)
		drainCount := 0
		for {
			select {
			case delivery, ok := <-msgs:
				if !ok {
					logger.Infof("queue %s drained successfully (%d messages)", queue, drainCount)
					return
				}
				atomic.AddInt32(&receivedMessagesCount, 1)
				drainCount++
				_ = delivery
				idleTimer.Reset(500 * time.Millisecond) // Reset idle timer on each message
			case <-idleTimer.C:
				logger.Infof("queue %s drain completed (no messages for 500ms, drained %d messages)", queue, drainCount)
				return
			}
		}
	}
}

func produceMessages(ch *amqp091.Channel, exchanges []string, exchangeRoutingKeys map[string][]string, messageSize int, duration time.Duration, doneChan <-chan struct{}) {
	timeout := time.After(duration)
	message := make([]byte, messageSize)
	rand.Read(message)

	// Enable publisher confirms for better performance tracking
	err := ch.Confirm(false)
	if err != nil {
		logger.Warnf("failed to enable publisher confirms: %v", err)
	}

	for {
		select {
		case <-doneChan:
			logger.Info("producer received stop signal, finishing...")
			return
		case <-timeout:
			logger.Info("producer timeout reached, stopping...")
			return
		default:
			exchange := exchanges[rand.Intn(len(exchanges))]
			routingKeys := exchangeRoutingKeys[exchange]
			if len(routingKeys) == 0 {
				continue
			}
			routingKey := routingKeys[rand.Intn(len(routingKeys))]
			err := ch.PublishWithContext(context.Background(), exchange, routingKey, false, false, amqp091.Publishing{
				Body: message,
			})
			if err != nil {
				logger.Errorf("failed to publish message to exchange %s with routing key %s: %v", exchange, routingKey, err)
			} else {
				atomic.AddInt32(&sentMessagesCount, 1)
			}
		}
	}
}
