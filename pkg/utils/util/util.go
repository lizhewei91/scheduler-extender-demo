package util

import (
	"path/filepath"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func BuildKubeConfig(configPath string) (*restclient.Config, error) {
	if len(configPath) != 0 {
		clientConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return BuildkubeConfigFromEnv()
		}
		return clientConfig, nil
	}

	return BuildkubeConfigFromEnv()
}

func BuildkubeConfigFromEnv() (*restclient.Config, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err == nil {
		return clientConfig, nil
	}

	return restclient.InClusterConfig()
}

func CommonElements(a, b []string) []string {
	// 创建一个 map 来存储 a 中的元素
	elementMap := make(map[string]bool)
	for _, item := range a {
		elementMap[item] = true
	}

	// 创建一个切片来存储相同的元素
	var common []string
	for _, item := range b {
		if elementMap[item] {
			common = append(common, item)
		}
	}

	return common
}
