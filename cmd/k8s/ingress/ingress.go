package ingress

import (
	"context"
	"fmt"
	"log"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	k8sutils "github.com/VojtechPastyrik/vp-utils/utils/k8s"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var Cmd = &cobra.Command{
	Use:     "ingress",
	Short:   "Ingress CLI Utils",
	Aliases: []string{"ing"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		listIngresses()
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func listIngresses() {
	// Use the same method to get the client as in the ArgoCD example
	clientset, _, err := k8sutils.KubernetesClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Print header
	fmt.Printf("%-20s %-30s %-50s\n", "NAMESPACE", "NAME", "HOSTS")

	// Get list of namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting namespaces: %v", err)
	}

	// Process all namespaces
	for _, ns := range namespaces.Items {
		listIngressesInNamespace(clientset, ns.Name)
	}

}

func listIngressesInNamespace(clientset *kubernetes.Clientset, namespace string) {
	ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting ingress in namespace %s: %v\n", namespace, err)
		return
	}

	for _, ing := range ingressList.Items {
		hosts := getIngressHosts(ing)
		fmt.Printf("%-20s %-30s %-50s\n", ing.Namespace, ing.Name, hosts)
	}
}

func getIngressHosts(ing networkingv1.Ingress) string {
	var hostsStr string

	// Map for storing TLS information for individual hosts
	tlsHosts := make(map[string]bool)

	// Fill the map of hosts with TLS configuration
	for _, tls := range ing.Spec.TLS {
		for _, host := range tls.Hosts {
			tlsHosts[host] = true
		}
	}

	for i, rule := range ing.Spec.Rules {
		if i > 0 {
			hostsStr += ","
		}

		// Check if host has TLS configuration
		protocol := "http"
		if tlsHosts[rule.Host] {
			protocol = "https"
		}

		hostsStr += protocol + "://" + rule.Host
	}
	return hostsStr
}
