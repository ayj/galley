// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	proxyconfig "istio.io/api/proxy/v1/config"
	"istio.io/galley/apiserver"
	"istio.io/galley/cmd"
	"istio.io/galley/model"
	"istio.io/galley/platform/kube"
)

type args struct {
	kubeconfig string
	meshConfig string

	ipAddress     string
	podName       string
	apiserverPort int

	// ingress sync mode is set to off by default
	controllerOptions kube.ControllerOptions
}

var (
	flags  args
	client *kube.Client
	mesh   *proxyconfig.ProxyMeshConfig

	rootCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "Legacy Istio apiserver",
		PersistentPreRunE: func(*cobra.Command, []string) (err error) {
			if flags.kubeconfig == "" {
				if v := os.Getenv("KUBECONFIG"); v != "" {
					glog.V(2).Infof("Setting configuration from KUBECONFIG environment variable")
					flags.kubeconfig = v
				}
			}

			client, err = kube.NewClient(flags.kubeconfig, model.IstioConfig)
			if err != nil {
				return multierror.Prefix(err, "failed to connect to Kubernetes API.")
			}
			if err = client.RegisterResources(); err != nil {
				return multierror.Prefix(err, "failed to register Third-Party Resources.")
			}

			// set values from environment variables
			if flags.ipAddress == "" {
				flags.ipAddress = os.Getenv("POD_IP")
			}
			if flags.podName == "" {
				flags.podName = os.Getenv("POD_NAME")
			}
			if flags.controllerOptions.Namespace == "" {
				flags.controllerOptions.Namespace = os.Getenv("POD_NAMESPACE")
			}
			glog.V(2).Infof("flags %s", spew.Sdump(flags))

			// receive mesh configuration
			mesh, err = cmd.GetMeshConfig(client.GetKubernetesClient(), flags.controllerOptions.Namespace, flags.meshConfig)
			if err != nil {
				return multierror.Prefix(err, "failed to retrieve mesh configuration.")
			}

			glog.V(2).Infof("mesh configuration %s", spew.Sdump(mesh))
			return
		},
	}

	apiserverCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "Start legacy Istio apiserver",
		Run: func(*cobra.Command, []string) {
			controller := kube.NewController(client, mesh, flags.controllerOptions)
			apiserver := apiserver.NewAPI(apiserver.APIServiceOptions{
				Version:  kube.IstioResourceVersion,
				Port:     flags.apiserverPort,
				Registry: &model.IstioRegistry{ConfigRegistry: controller},
			})
			stop := make(chan struct{})
			go controller.Run(stop)
			go apiserver.Run()
			cmd.WaitSignal(stop)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.kubeconfig, "kubeconfig", "",
		"Use a Kubernetes configuration file instead of in-cluster configuration")
	rootCmd.PersistentFlags().StringVarP(&flags.controllerOptions.Namespace, "namespace", "n", "",
		"Select a namespace for the controller loop. If not set, uses ${POD_NAMESPACE} environment variable")
	rootCmd.PersistentFlags().DurationVar(&flags.controllerOptions.ResyncPeriod, "resync", time.Second,
		"Controller resync interval")
	rootCmd.PersistentFlags().StringVar(&flags.controllerOptions.DomainSuffix, "domainSuffix", "cluster.local",
		"Kubernetes DNS domain suffix")
	rootCmd.PersistentFlags().StringVar(&flags.meshConfig, "meshConfig", cmd.DefaultConfigMapName,
		fmt.Sprintf("ConfigMap name for Istio mesh configuration, config key should be %q", cmd.ConfigMapKey))

	apiserverCmd.PersistentFlags().IntVar(&flags.apiserverPort, "port", 8081,
		"Config API service port")

	cmd.AddFlags(rootCmd)

	rootCmd.AddCommand(apiserverCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		glog.Error(err)
		os.Exit(-1)
	}
}
