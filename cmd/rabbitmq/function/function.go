package function

import (
	"fmt"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/rabbitmq"
	rabbitmqUtisl "github.com/VojtechPastyrik/vpd/utils/rabbitmq"
	"github.com/pterm/pterm"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var (
	FlagHost        string
	FlagPort        int
	FlagUser        string
	FlagPassword    string
	FlagVirtualHost string
	FlagExchange    string
	FlagQueue       string
	FlagRoutingKey  string
	FlagSsl         bool
	FlagSslCert     string
	FlagSslKey      string
)

var Cmd = &cobra.Command{
	Use:     "func-test",
	Aliases: []string{"ft"},
	Short:   "Run functional test on RabbitMQ server",
	Long:    "Run functional test on RabbitMQ server. It will connect to the server, create exchange, queue and binding, send a message, read it back and clean up after the test.",
	Run: func(cmd *cobra.Command, args []string) {
		runFunctionalTest(FlagHost, FlagPort, FlagUser, FlagPassword, FlagVirtualHost, FlagExchange, FlagQueue, FlagRoutingKey, FlagSsl, FlagSslCert, FlagSslKey)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagHost, "host", "H", "localhost", "RabbitMQ host")
	Cmd.MarkFlagRequired("host")
	Cmd.Flags().IntVarP(&FlagPort, "port", "P", 5672, "RabbitMQ port")
	Cmd.MarkFlagRequired("port")
	Cmd.Flags().StringVarP(&FlagUser, "user", "u", "guest", "RabbitMQ user")
	Cmd.MarkFlagRequired("user")
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "guest", "RabbitMQ password")
	Cmd.MarkFlagRequired("password")
	Cmd.Flags().StringVarP(&FlagVirtualHost, "vhost", "v", "/", "RabbitMQ virtual host")
	Cmd.Flags().StringVarP(&FlagExchange, "exchange", "e", "", "RabbitMQ exchange name to use for testing")
	Cmd.Flags().StringVarP(&FlagQueue, "queue", "q", "", "RabbitMQ queue name to use for testing")
	Cmd.Flags().StringVarP(&FlagRoutingKey, "routing-key", "r", "", "RabbitMQ routing key to use for testing")
	Cmd.Flags().BoolVarP(&FlagSsl, "ssl", "s", false, "Use SSL for RabbitMQ connection")
	Cmd.Flags().StringVarP(&FlagSslCert, "ssl-cert", "c", "", "Path to SSL certificate file")
	Cmd.Flags().StringVarP(&FlagSslKey, "ssl-key", "k", "", "Path to SSL key file")
}

func runFunctionalTest(host string, port int, user, password, virtualHost, exchange, queue, routingKey string, ssl bool, sslCert, sslKey string) {
	if exchange == "" && (queue != "" || routingKey != "") {

		pterm.Warning.Println("If one of 'exchange', 'queue', or 'routingKey' is set, all must be set. Only direct exchange can be used for functional testing.")
		return
	}
	exchangeUsed := exchange
	if exchangeUsed == "" {
		exchangeUsed = "rabbitmq-function-test-exchange"
	}
	queueUsed := queue
	if queueUsed == "" {
		queueUsed = "rabbitmq-function-test-queue"
	}
	routingKeyUsed := routingKey
	if routingKeyUsed == "" {
		routingKeyUsed = "rabbitmq-function-test-routing-key"
	}

	// Task: Connect to RabbitMQ
	spinnerConnect, _ := pterm.DefaultSpinner.Start("Running Connect to RabbitMQ...")
	con, ch, err := rabbitmqUtisl.ConnectToRabbitMQ(ssl, user, password, host, port, virtualHost, sslCert, sslKey)
	if err != nil {
		spinnerConnect.Fail(fmt.Sprintf("Connection to RabbitMQ failed: %s", err.Error()))
		return
	}
	spinnerConnect.Success("Connected to RabbitMQ")
	defer func() {
		if err := con.Close(); err != nil {
			pterm.Error.Printf("Error closing RabbitMQ connection: %s\n", err)
		}
	}()
	defer func() {
		if err := ch.Close(); err != nil {
			pterm.Error.Printf("Error closing RabbitMQ channel: %s\n", err)
		}
	}()

	// Task: Create exchange, binding and queue
	spinnerCreate, _ := pterm.DefaultSpinner.Start("Create exchange, queue and binding...")
	if exchange == "" {
		err := ch.ExchangeDeclare(exchangeUsed, "direct", true, false, false, false, nil)
		if err != nil {
			spinnerCreate.Fail(fmt.Sprintf("Failed to create exchange: %s", err.Error()))
			return
		}
		_, err = ch.QueueDeclare(queueUsed, true, false, false, false, nil)
		if err != nil {
			spinnerCreate.Fail(fmt.Sprintf("Failed to create queue: %s", err.Error()))
			return
		}
		err = ch.QueueBind(queueUsed, routingKeyUsed, exchangeUsed, false, nil)
		if err != nil {
			spinnerCreate.Fail(fmt.Sprintf("Failed to bind queue to exchange: %s", err.Error()))
			return
		}
		spinnerCreate.Success("All objects created successfully")
	} else {
		spinnerCreate.Info("All parameters are specified, skipping objects creation")
	}

	// Task: Send message
	spinnerSend, _ := pterm.DefaultSpinner.Start("Sending message to RabbitMQ...")
	sendErr := ch.Publish(
		exchangeUsed,
		routingKeyUsed,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"message": "This is a test message"}`),
		},
	)
	if sendErr != nil {
		spinnerSend.Fail(fmt.Sprintf("Failed to publish message: %s", sendErr.Error()))
		return
	}
	spinnerSend.Success("Message sent successfully")

	// Task: Read message
	spinnerRead, _ := pterm.DefaultSpinner.Start("Reading message from RabbitMQ...")
	msgs, err := ch.Consume(
		queueUsed,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		spinnerRead.Fail(fmt.Sprintf("Failed to consume messages: %s", err.Error()))
		return
	}
	select {
	case _ = <-msgs:
		// IMPORTANT: Complete the spinner BEFORE logging the message
		spinnerRead.Success("Message received successfully")
	case <-time.After(5 * time.Second):
		spinnerRead.Fail("No message received within 5 seconds timeout")
	}

	// Task: Clean after the test
	spinnerCleanup, _ := pterm.DefaultSpinner.Start("Cleaning up after the test...")
	if exchange == "" {
		if err := ch.QueueUnbind(queueUsed, routingKeyUsed, exchangeUsed, nil); err != nil {
			pterm.Error.Printf("Failed to unbind queue: %s\n", err)
		} else {
			pterm.Info.Print("Successfully unbound queue\n")
		}

		if _, err := ch.QueueDelete(queueUsed, false, false, false); err != nil {
			pterm.Error.Printf("Failed to delete queue: %s\n", err)

		} else {
			pterm.Info.Print("Successfully deleted queue\n")
		}

		if err := ch.ExchangeDelete(exchangeUsed, false, false); err != nil {
			pterm.Error.Printf("Failed to delete exchange: %s\n", err)
		} else {
			pterm.Info.Print("Successfully deleted exchange\n")
		}
		spinnerCleanup.Success("Cleaned up after the test")
	} else {
		spinnerCleanup.Info("Parameters were specified, skipping object deletion.")
	}
}
