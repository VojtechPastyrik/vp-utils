package load

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/rabbitmq"
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
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
		log.Printf("Connection to RabbitMQ failed: %s", err.Error())
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
		log.Println("Signal received, cleaning up resources...")
		cleanupResources(ch, exchangeList, queueList)
		os.Exit(0)
	}()

	durationParsed, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("Invalid duration format: %s\n", err.Error())
		return
	}

	var wg sync.WaitGroup

	startConsumersAndProducers(con, queueList, exchangeList, exchangeRoutingKeys, parallelClients, durationParsed, messageSize)

	wg.Wait()

	log.Printf("Load test completed. Sent messages: %d, Received messages: %d\n", sentMessagesCount, receivedMessagesCount)
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
			log.Printf("Failed to declare exchange %s: %s", exchangeName, err.Error())
		}
		exchangeList[i] = exchangeName
	}

	for i := 0; i < queueCount; i++ {
		queueName := fmt.Sprintf("test-queue-%d", i)
		_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			log.Printf("Failed to declare queue %s: %s", queueName, err.Error())
		}
		queueList[i] = queueName
	}

	for i := 0; i < routingKeyCount; i++ {
		routingKeys[i] = fmt.Sprintf("test-routing-key-%d", i)
	}

	for _, queue := range queueList {
		routingKey := routingKeys[rand.Intn(routingKeyCount)]
		exchange := exchangeList[rand.Intn(exchangeCount)]
		err := ch.QueueBind(queue, routingKey, exchange, false, nil)
		if err != nil {
			log.Printf("Failed to bind queue %s to exchange %s with routing key %s: %s", queue, exchange, routingKey, err.Error())
		}
		exchangeRoutingKeys[exchange] = append(exchangeRoutingKeys[exchange], routingKey)
	}

	for _, routingKey := range routingKeys {
		for i := 0; i < 3; i++ {
			queue := queueList[rand.Intn(queueCount)]
			exchange := exchangeList[rand.Intn(exchangeCount)]
			err := ch.QueueBind(queue, routingKey, exchange, false, nil)
			if err != nil {
				log.Printf("Failed to bind queue %s to exchange %s with routing key %s: %s", queue, exchange, routingKey, err.Error())
			}
			exchangeRoutingKeys[exchange] = append(exchangeRoutingKeys[exchange], routingKey)
		}
	}

	return exchangeList, queueList, exchangeRoutingKeys
}

func cleanupResources(ch *amqp091.Channel, exchangeList, queueList []string) {
	for _, exchange := range exchangeList {
		err := ch.ExchangeDelete(exchange, false, false)
		if err != nil {
			log.Printf("Failed to delete exchange %s: %s", exchange, err.Error())
		}
	}

	for _, queue := range queueList {
		_, err := ch.QueueDelete(queue, false, false, false)
		if err != nil {
			log.Printf("Failed to delete queue %s: %s", queue, err.Error())
		}
	}
}

func startConsumersAndProducers(conn *amqp091.Connection, queueList, exchangeList []string, exchangeRoutingKeys map[string][]string, parallelClients int, duration time.Duration, messageSize int) {
	var wg sync.WaitGroup
	doneChan := make(chan struct{})

	log.Println("Starting consumers...")
	assignedQueues := assignQueuesToConsumers(queueList, parallelClients)
	for i := 0; i < parallelClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			log.Printf("Consumer %d started.\n", id)
			ch, err := conn.Channel()
			if err != nil {
				log.Printf("Failed to create channel for consumer %d: %s\n", id, err.Error())
				return
			}
			defer ch.Close()

			log.Printf("Consumer %d assigned queues: %v\n", id, assignedQueues[id])
			consumeMessages(ch, assignedQueues[id], duration, doneChan)
			log.Printf("Consumer %d finished.\n", id)
		}(i)
	}

	log.Println("Starting producers...")
	for i := 0; i < parallelClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			log.Printf("Producer %d started.\n", id)
			ch, err := conn.Channel()
			if err != nil {
				log.Printf("Failed to create channel for producer %d: %s\n", id, err.Error())
				return
			}
			defer ch.Close()

			assignedExchange := exchangeList[rand.Intn(len(exchangeList))]
			assignedRoutingKeys := exchangeRoutingKeys[assignedExchange]
			log.Printf("Producer %d assigned exchange: %s and routing keys: %v\n", id, assignedExchange, assignedRoutingKeys)
			produceMessages(ch, []string{assignedExchange}, assignedRoutingKeys, messageSize, duration, doneChan)
			log.Printf("Producer %d finished.\n", id)
		}(i)
	}

	go func() {
		time.Sleep(duration)
		log.Println("Duration expired, stopping load test...")
		close(doneChan)
	}()

	log.Println("Waiting for all goroutines to finish...")
	wg.Wait()
	log.Println("All producers and consumers stopped.")
}

func assignQueuesToConsumers(queueList []string, parallelClients int) [][]string {
	assigned := make([][]string, parallelClients)
	for i, queue := range queueList {
		assigned[i%parallelClients] = append(assigned[i%parallelClients], queue)
	}
	return assigned
}

func consumeMessages(ch *amqp091.Channel, queues []string, duration time.Duration, doneChan <-chan struct{}) {
	var wg sync.WaitGroup
	timeout := time.After(duration)

	for _, queue := range queues {
		wg.Add(1)
		go func(queue string) {
			defer wg.Done()
			for {
				select {
				case <-doneChan:
					log.Printf("Consumer received stop signal for queue %s, exiting...\n", queue)
					drainQueues(ch, []string{queue}) // Vyprázdnění fronty při ukončení konzumenta
					return
				case <-timeout:
					log.Printf("Consumer timeout reached for queue %s, exiting...\n", queue)
					drainQueues(ch, []string{queue}) // Vyprázdnění fronty při dosažení timeoutu
					return
				default:
					_, ok, err := ch.Get(queue, true)
					if err != nil {
						log.Printf("Failed to consume message from queue %s: %s\n", queue, err.Error())
						continue
					}
					if ok {
						atomic.AddInt32(&receivedMessagesCount, 1)
					}
				}
			}
		}(queue)
	}

	wg.Wait()
}

func drainQueues(ch *amqp091.Channel, queues []string) {
	for _, queue := range queues {
		for {
			qInfo, err := ch.QueueInspect(queue)
			if err != nil {
				log.Printf("Failed to inspect queue %s: %s\n", queue, err.Error())
				break
			}
			if qInfo.Messages == 0 {
				log.Printf("Queue %s is empty, stopping consumer.\n", queue)
				break
			}
			for qInfo.Messages > 0 {
				_, ok, err := ch.Get(queue, true)
				if err != nil {
					log.Printf("Failed to consume message from queue %s: %s\n", queue, err.Error())
					break
				}
				if ok {
					atomic.AddInt32(&receivedMessagesCount, 1)
				}
				qInfo, err = ch.QueueInspect(queue)
				if err != nil {
					log.Printf("Failed to inspect queue %s: %s\n", queue, err.Error())
					break
				}
			}
		}
	}
}

func produceMessages(ch *amqp091.Channel, exchanges []string, routingKeys []string, messageSize int, duration time.Duration, doneChan <-chan struct{}) {
	timeout := time.After(duration)
	message := make([]byte, messageSize)
	rand.Read(message)

	for {
		select {
		case <-doneChan:
			log.Println("Producer received stop signal, finishing...")
			return
		case <-timeout:
			log.Println("Producer timeout reached, stopping...")
			return
		default:
			exchange := exchanges[rand.Intn(len(exchanges))]
			routingKey := routingKeys[rand.Intn(len(routingKeys))]
			err := ch.Publish(exchange, routingKey, false, false, amqp091.Publishing{
				Body: message,
			})
			if err != nil {
				log.Printf("Failed to publish message to exchange %s with routing key %s: %s\n", exchange, routingKey, err.Error())
			} else {
				atomic.AddInt32(&sentMessagesCount, 1) // Atomické zvýšení počítadla
			}
		}
	}
}
