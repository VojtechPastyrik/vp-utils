package subscription

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/subscriptions"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/azure"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	FlagChangeSub        bool
	FlagAllSubscriptions bool
)

var Cmd = &cobra.Command{
	Use:     "subscription [flags] <string>",
	Short:   "Subscription commands",
	Long:    `This command provides various tools for managing Azure subscriptions, such as a better view of the subscription listing and switching between subscriptions`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"sub"},
	Run: func(cmd *cobra.Command, args []string) {
		subName := ""
		if len(args) > 0 {
			subName = args[0]
		}
		subscription(FlagChangeSub, FlagAllSubscriptions, subName)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagChangeSub,
		"change",
		"c", false,
		"Change the current subscription to the one specified by the string argument. If the string is empty, it will list all subscriptions and prompt for selection.")
	Cmd.Flags().BoolVarP(&FlagAllSubscriptions,
		"all",
		"a",
		false,
		"List all subscriptions, not just the active ones or just for logged-in user. This is useful for administrators who need to manage multiple subscriptions across different users or tenants.")
}

func subscription(changeSub, allSubs bool, arg string) {
	// Use environment variables for authentication
	authorizer, err := auth.NewAuthorizerFromCLI()
	handleError(err)

	// Create a client
	subscriptionsClient := subscriptions.NewClient()
	subscriptionsClient.Authorizer = authorizer

	if !changeSub {
		listSubscriptions(subscriptionsClient, allSubs)
		return
	} else {
		if arg == "" {
			// If no argument is provided, list subscriptions and prompt for selection
			listSubscriptions(subscriptionsClient, allSubs)
			fmt.Print("Enter the subscription ID to switch to: ")
			fmt.Scanln(&arg)
		}

		// Change the subscription
		cmd := exec.Command("az", "account", "set", "--subscription", arg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			logger.Fatalf("error changing subscription: %v", err)
		} else {
			fmt.Printf("Switched to subscription: %s\n", arg)

		}
	}
}

func listSubscriptions(subscriptionsClient subscriptions.Client, all bool) {
	//GET active subscription ID
	cmdActive := exec.Command("az", "account", "show", "--query", "id", "-o", "tsv")
	activeID, err := cmdActive.Output()
	if err != nil {
		logger.Fatalf("cannot get active subscription: %v", err)
		activeID = []byte("")
	}
	activeIDStr := strings.TrimSpace(string(activeID))

	var currentUser string
	if !all {
		cmdUser := exec.Command("az", "account", "show", "--query", "user.name", "-o", "tsv")
		userData, err := cmdUser.Output()
		if err != nil {
			logger.Infof("warning: failed to get current user name: %v", err)
		} else {
			currentUser = strings.TrimSpace(string(userData))
		}
	}

	// Construct the query for listing subscriptions
	var query string
	if !all && currentUser != "" {
		// Filter subscriptions by the current user
		query = fmt.Sprintf("[?user.name=='%s'].{Name:name, SubscriptionId:id, TenantId:tenantId, User:user.name, State:state, Active: (id=='%s') && '[X]' || ''}",
			currentUser, activeIDStr)
	} else {
		// List all subscriptions without filtering by user
		query = fmt.Sprintf("[].{Name:name, SubscriptionId:id, TenantId:tenantId, User:user.name, State:state, Active: (id=='%s') && '[X]' || ''}", activeIDStr)
	}

	// Run the command to list subscriptions
	cmd := exec.Command("az", "account", "list", "--query", query, "--output", "table")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		logger.Fatalf("error listing subscriptions: %v", err)
	}
}

func handleError(err error) {
	if err != nil {
		logger.Fatalf("%v", err)
	}
}
