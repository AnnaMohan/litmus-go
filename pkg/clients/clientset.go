package clients

import (
	"flag"

	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
)

// ClientSets is a collection of clientSets and kubeConfig needed
type ClientSets struct {
	KubeClient    *kubernetes.Clientset
	LitmusClient  *chaosClient.LitmuschaosV1alpha1Client
	KubeConfig    *rest.Config
	DynamicClient dynamic.Interface
}

// GenerateClientSetFromKubeConfig will generation both ClientSets (k8s, and Litmus) as well as the KubeConfig
func (clientSets *ClientSets) GenerateClientSetFromKubeConfig() error {

	config, err := getKubeConfig()
	if err != nil {
		return err
	}
	k8sClientSet, err := generateK8sClientSet(config)
	if err != nil {
		return err
	}
	litmusClientSet, err := generateLitmusClientSet(config)
	if err != nil {
		return err
	}
	dynamicClientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	clientSets.KubeClient = k8sClientSet
	clientSets.LitmusClient = litmusClientSet
	clientSets.KubeConfig = config
	clientSets.DynamicClient = dynamicClientSet
	return nil
}

// getKubeConfig setup the config for access cluster resource
func getKubeConfig() (*rest.Config, error) {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	// It uses in-cluster config, if kubeconfig path is not specified
	config, err := buildConfigFromFlags("", *kubeconfig)
	return config, err
}

// generateK8sClientSet will generation k8s client
func generateK8sClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to generate kubernetes clientSet, err: %v: ", err)
	}
	return k8sClientSet, nil
}

// generateLitmusClientSet will generate a LitmusClient
func generateLitmusClientSet(config *rest.Config) (*chaosClient.LitmuschaosV1alpha1Client, error) {
	litmusClientSet, err := chaosClient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create LitmusClientSet, err: %v", err)
	}
	return litmusClientSet, nil
}

// buildConfigFromFlags is a helper function that builds configs from a master
// url or a kubeconfig filepath, if nothing is provided it falls back to inClusterConfig
func buildConfigFromFlags(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
	if kubeconfigPath == "" && masterUrl == "" {
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		klog.Warningf("Neither --kubeconfig nor --master was specified.  Using the inClusterConfig. Error creating inClusterConfig: ", err)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientConfig()
}
