package main

import (
	"context"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// This func returns the list of files that needs to be updated from a target directory
func findFile(targetDir string, pattern []string) []string {
	var patternMatch []string
	for _, v := range pattern {
		matches, err := filepath.Glob(targetDir + v)

		if err != nil {
			fmt.Println(err)
		}

		if len(matches) != 0 {
			patternMatch = append(patternMatch, matches...)
		}
	}
	return patternMatch
}

// Converts the file data into map that can be consumed to create a k8s resource
func cmdatamap(filelist []string) map[string]string {

	cmdata := make(map[string]string)
	for _, v := range filelist {
		v1 := strings.Split(v, "/")

		data, err := ioutil.ReadFile(v)

		if err != nil {
			log.Panicf("failed reading data from file: %s", err)
		} else {
			cmdata[v1[len(v1)-1]] = string(data)
		}

	}

	return cmdata
}
func main() {

	//Directory in which the config files are present
	targetDirectory := "/Users/vasudev.akkur/src/scalyr-config/config/"

	//Takes the list of files names that needs to updated in configmap
	fileName := []string{"worker", "atlantis"}

	//Generates a slice string on configs that needs to be updated.
	configfileslist := findFile(targetDirectory, fileName)
	fmt.Println(configfileslist)

	// This is test but we run within k8s cluster we can use https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go#L41-L50
	kubeconfig := os.Getenv("HOME") + "/.kube/config"

	// create the config object from kubeconfig
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// create clientset
	clientset, _ := kubernetes.NewForConfig(config)

	//Map of the config data
	datastringmap := cmdatamap(configfileslist)

	// executes GET request to K8s API and gets the current configmap from the cluster
	getCM, _ := clientset.CoreV1().ConfigMaps("default").Get(context.Background(), "test-config", metav1.GetOptions{})

	// V1 api takes the configuration with which the configmap needs to be created.
	k8sconfigmap := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "default"}, Data: datastringmap}

	//Checks if the CM exists, if it does not exist it will create or it will update
	if getCM.ObjectMeta.Name == "" {
		CreateCM, err := clientset.CoreV1().ConfigMaps("default").Create(context.Background(), k8sconfigmap, metav1.CreateOptions{})
		if err != nil {
			log.Fatal("Could not create the configmap, check the file configuration")
			os.Exit(1)
		}

		fmt.Println("The configmap is created \n", CreateCM.ObjectMeta)
	} else {
		updateCM, err := clientset.CoreV1().ConfigMaps("default").Update(context.Background(), k8sconfigmap, metav1.UpdateOptions{})

		if err != nil {
			log.Fatal("Could not create the configmap, check the file configuration")
			os.Exit(1)
		}

		fmt.Println("The configmap is updated \n", updateCM.ObjectMeta)
	}

}
