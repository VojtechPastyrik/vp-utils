package load

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/rabbitmq"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	rabbitmqUtisl "github.com/VojtechPastyrik/vpd/utils/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

type LoadProfile struct {
	Duration        string
	QueueCount      int
	ExchangeCount   int
	RoutingKeyCount int
	MessageSize     int
	ParallelClients int
	Description     string
}

var loadProfiles = map[string]LoadProfile{
	"light": {
		Duration:        "30s",
		QueueCount:      5,
		ExchangeCount:   2,
		RoutingKeyCount: 3,
		MessageSize:     1024,
		ParallelClients: 2,
		Description:     "Light load - sanity check (2.5K-5K msgs/sec expected)",
	},
	"medium": {
		Duration:        "2m",
		QueueCount:      20,
		ExchangeCount:   5,
		RoutingKeyCount: 10,
		MessageSize:     4096,
		ParallelClients: 10,
		Description:     "Medium load - typical workload (20K-50K msgs/sec expected)",
	},
	"heavy": {
		Duration:        "3m",
		QueueCount:      50,
		ExchangeCount:   20,
		RoutingKeyCount: 30,
		MessageSize:     8192,
		ParallelClients: 20,
		Description:     "Heavy load - stress test (100K+ msgs/sec expected)",
	},
	"sustained": {
		Duration:        "10m",
		QueueCount:      30,
		ExchangeCount:   10,
		RoutingKeyCount: 15,
		MessageSize:     2048,
		ParallelClients: 15,
		Description:     "Sustained load - stability test (long-running)",
	},
}

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
	FlagLoadProfile       string
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
		runLoadTestWithProfile(FlagHost, FlagPort, FlagUser, FlagPassword, FlagVirtualHost, FlagSsl, FlagSslCert, FlagSslKey, FlagLoadProfile, FlagDuration, FlagQueueCount, FlagExchangeCount, FlagRoutingKeyCount, FlagMessageSize, FlagParallelClients)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagHost, "host", "H", "localhost", "RabbitMQ host")
	Cmd.Flags().IntVarP(&FlagPort, "port", "P", 5672, "RabbitMQ port")
	Cmd.Flags().StringVarP(&FlagUser, "user", "u", "guest", "RabbitMQ username")
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "guest", "RabbitMQ password")
	Cmd.Flags().StringVarP(&FlagVirtualHost, "vhost", "v", "/", "RabbitMQ virtual host")
	Cmd.Flags().BoolVarP(&FlagSsl, "ssl", "s", false, "Enable SSL")
	Cmd.Flags().StringVarP(&FlagSslCert, "ssl-cert", "c", "", "Path to SSL certificate")
	Cmd.Flags().StringVarP(&FlagSslKey, "ssl-key", "k", "", "Path to SSL key")
	Cmd.Flags().StringVarP(&FlagLoadProfile, "profile", "L", "medium", "Load profile: light, medium, heavy, sustained (overrides individual settings)")
	Cmd.Flags().StringVarP(&FlagDuration, "duration", "d", "", "Duration of the load test (overrides profile)")
	Cmd.Flags().IntVarP(&FlagQueueCount, "queue-count", "q", 0, "Number of queues to create (overrides profile)")
	Cmd.Flags().IntVarP(&FlagExchangeCount, "exchange-count", "e", 0, "Number of exchanges to create (overrides profile)")
	Cmd.Flags().IntVarP(&FlagRoutingKeyCount, "routing-keys", "r", 0, "Number of routing keys to use (overrides profile)")
	Cmd.Flags().IntVarP(&FlagMessageSize, "message-size", "m", 0, "Size of each message in bytes (overrides profile)")
	Cmd.Flags().IntVarP(&FlagParallelClients, "parallel-clients", "C", 0, "Number of parallel clients (overrides profile)")
}

func runLoadTestWithProfile(host string, port int, user, password, virtualHost string, ssl bool, sslCert, sslKey, profile, duration string, queueCount, exchangeCount, routingKeyCount, messageSize, parallelClients int) {
	// Load profile settings
	loadProfile, exists := loadProfiles[profile]
	if !exists {
		logger.Errorf("invalid profile '%s'. Available profiles: light, medium, heavy, sustained", profile)
		return
	}

	// Apply profile defaults, then override with explicit flags
	finalDuration := duration
	if finalDuration == "" {
		finalDuration = loadProfile.Duration
	}

	finalQueueCount := queueCount
	if finalQueueCount == 0 {
		finalQueueCount = loadProfile.QueueCount
	}

	finalExchangeCount := exchangeCount
	if finalExchangeCount == 0 {
		finalExchangeCount = loadProfile.ExchangeCount
	}

	finalRoutingKeyCount := routingKeyCount
	if finalRoutingKeyCount == 0 {
		finalRoutingKeyCount = loadProfile.RoutingKeyCount
	}

	finalMessageSize := messageSize
	if finalMessageSize == 0 {
		finalMessageSize = loadProfile.MessageSize
	}

	finalParallelClients := parallelClients
	if finalParallelClients == 0 {
		finalParallelClients = loadProfile.ParallelClients
	}

	logger.Infof("using load profile: %s", profile)
	logger.Infof("%s", loadProfile.Description)

	runLoadTest(host, port, user, password, virtualHost, ssl, sslCert, sslKey, finalDuration, finalQueueCount, finalExchangeCount, finalRoutingKeyCount, finalMessageSize, finalParallelClients)
}

func runLoadTest(host string, port int, user, password, virtualHost string, ssl bool, sslCert, sslKey, duration string, queueCount, exchangeCount, routingKeyCount, messageSize, parallelClients int) {
	startTime := time.Now()

	con, ch, err := rabbitmqUtisl.ConnectToRabbitMQ(ssl, user, password, host, port, virtualHost, sslCert, sslKey)
	if err != nil {
		logger.Errorf("connection to rabbitmq failed: %v", err)
		return
	}
	defer con.Close()

	exchangeList, queueList, exchangeRoutingKeys := createResources(ch, exchangeCount, queueCount, routingKeyCount)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	durationParsed, err := time.ParseDuration(duration)
	if err != nil {
		logger.Errorf("invalid duration format: %v", err)
		return
	}

	// Run test in a goroutine so we can handle signals
	done := make(chan struct{})
	go func() {
		startConsumersAndProducers(con, queueList, exchangeList, exchangeRoutingKeys, parallelClients, durationParsed, messageSize)
		close(done)
	}()

	// Wait for either test to complete or signal
	select {
	case <-done:
		logger.Info("load test completed normally")
	case sig := <-sigChan:
		logger.Infof("signal received (%v), stopping test...", sig)
	}

	endTime := time.Now()
	printStatistics(startTime, endTime, sentMessagesCount, receivedMessagesCount, queueCount, exchangeCount, routingKeyCount, messageSize, parallelClients, durationParsed)

	logger.Info("starting cleanup...")
	cleanupResources(ch, exchangeList, queueList)
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
		// Create quorum queue by setting x-queue-type to "quorum"
		args := amqp091.Table{
			"x-queue-type": "quorum",
		}
		_, err := ch.QueueDeclare(queueName, true, false, false, false, args)
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
	// First, purge all queues to remove remaining messages
	for _, queue := range queueList {
		purgedCount, err := ch.QueuePurge(queue, false)
		if err != nil {
			logger.Errorf("failed to purge queue %s: %v", queue, err)
		} else {
			logger.Infof("purged %d messages from queue %s", purgedCount, queue)
		}
	}

	// Unbind queues from exchanges
	for _, exchange := range exchangeList {
		for _, queue := range queueList {
			err := ch.QueueUnbind(queue, "", exchange, nil)
			if err != nil {
				logger.Debugf("failed to unbind queue %s from exchange %s: %v", queue, exchange, err)
				// Continue even if unbind fails - queue might not be bound with this routing key
			}
		}
	}

	// Delete exchanges
	for _, exchange := range exchangeList {
		err := ch.ExchangeDelete(exchange, false, false)
		if err != nil {
			logger.Errorf("failed to delete exchange %s: %v", exchange, err)
		} else {
			logger.Infof("deleted exchange %s", exchange)
		}
	}

	// Delete queues
	for _, queue := range queueList {
		deletedCount, err := ch.QueueDelete(queue, false, false, false)
		if err != nil {
			logger.Errorf("failed to delete queue %s: %v", queue, err)
		} else {
			logger.Infof("deleted queue %s (%d messages remaining)", queue, deletedCount)
		}
	}

	logger.Info("cleanup completed")
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
	if len(exchanges) == 0 {
		logger.Warnf("no exchanges provided, producer exiting")
		return
	}

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
			if len(exchanges) == 0 {
				continue
			}
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

func printStatistics(startTime, endTime time.Time, sentMessages, receivedMessages int32, queueCount, exchangeCount, routingKeyCount, messageSize, parallelClients int, expectedDuration time.Duration) {
	actualDuration := endTime.Sub(startTime)
	durationSeconds := actualDuration.Seconds()
	expectedSeconds := expectedDuration.Seconds()

	sentMsgsPerSecond := float64(sentMessages) / durationSeconds
	receivedMsgsPerSecond := float64(receivedMessages) / durationSeconds

	totalDataSent := int64(sentMessages) * int64(messageSize)
	totalDataReceived := int64(receivedMessages) * int64(messageSize)

	dataSentMBps := float64(totalDataSent) / (1024 * 1024 * durationSeconds)
	dataReceivedMBps := float64(totalDataReceived) / (1024 * 1024 * durationSeconds)

	duplicationFactor := float64(receivedMessages) / float64(sentMessages)
	timeMultiplier := durationSeconds / expectedSeconds

	separator := strings.Repeat("=", 70)
	fmt.Println("\n" + separator)
	fmt.Println("                    LOAD TEST STATISTICS                         ")
	fmt.Println(separator)
	fmt.Printf("\nTest Configuration:\n")
	fmt.Printf("  - Expected Duration:     %.2f seconds\n", expectedSeconds)
	fmt.Printf("  - Actual Duration:       %.2f seconds\n", durationSeconds)
	fmt.Printf("  - Time Multiplier:       %.2f x\n", timeMultiplier)
	fmt.Printf("  - Queues:                %d (type: quorum)\n", queueCount)
	fmt.Printf("  - Exchanges:             %d\n", exchangeCount)
	fmt.Printf("  - Routing Keys:          %d\n", routingKeyCount)
	fmt.Printf("  - Message Size:          %d bytes\n", messageSize)
	fmt.Printf("  - Parallel Clients:      %d (consumers + producers)\n", parallelClients)

	fmt.Printf("\nMessage Statistics:\n")
	fmt.Printf("  - Messages Sent:         %d\n", sentMessages)
	fmt.Printf("  - Messages Received:     %d\n", receivedMessages)
	fmt.Printf("  - Duplication Factor:    %.2f x\n", duplicationFactor)

	fmt.Printf("\nThroughput:\n")
	fmt.Printf("  - Send Rate:             %.2f msgs/sec\n", sentMsgsPerSecond)
	fmt.Printf("  - Receive Rate:          %.2f msgs/sec\n", receivedMsgsPerSecond)

	fmt.Printf("\nData Transfer:\n")
	fmt.Printf("  - Total Data Sent:       %.2f MB\n", float64(totalDataSent)/(1024*1024))
	fmt.Printf("  - Total Data Received:   %.2f MB\n", float64(totalDataReceived)/(1024*1024))
	fmt.Printf("  - Send Rate:             %.2f MB/s\n", dataSentMBps)
	fmt.Printf("  - Receive Rate:          %.2f MB/s\n", dataReceivedMBps)

	fmt.Println("\n" + separator)
}
