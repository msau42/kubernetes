/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// This file implements a dummy scheduler extender.

import (
	"encoding/json"
	"flag"
	"net/http"
	"strings"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
)

const (
	filter     = "filter"
	prioritize = "prioritize"
	bind       = "bind"
)

type fitPredicate func(pod *v1.Pod, node *v1.Node) (bool, error)
type priorityFunc func(pod *v1.Pod, nodes *v1.NodeList) (*schedulerapi.HostPriorityList, error)

type priorityConfig struct {
	function priorityFunc
	weight   int
}

type Extender struct {
	name             string
	predicates       []fitPredicate
	prioritizers     []priorityConfig
	nodeCacheCapable bool
	Client           clientset.Interface
}

func (e *Extender) serveHTTP(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	encoder := json.NewEncoder(w)

	if strings.Contains(req.URL.Path, filter) || strings.Contains(req.URL.Path, prioritize) {
		var args schedulerapi.ExtenderArgs

		if err := decoder.Decode(&args); err != nil {
			http.Error(w, "Decode error", http.StatusBadRequest)
			return
		}

		if strings.Contains(req.URL.Path, filter) {
			glog.Infof("Got filter request")
			resp := &schedulerapi.ExtenderFilterResult{}
			resp, err := e.Filter(&args)
			if err != nil {
				resp.Error = err.Error()
			}

			if err := encoder.Encode(resp); err != nil {
				glog.Fatalf("Failed to encode %v", resp)
			}
		} else if strings.Contains(req.URL.Path, prioritize) {
			glog.Infof("Got prioritize request")
			// Prioritize errors are ignored. Default k8s priorities or another extender's
			// priorities may be applied.
			priorities, _ := e.Prioritize(&args)

			if err := encoder.Encode(priorities); err != nil {
				glog.Fatalf("Failed to encode %+v", priorities)
			}
		}
	} else {
		http.Error(w, "Unknown method", http.StatusNotFound)
	}
}

func (e *Extender) Filter(args *schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	failedNodesMap := schedulerapi.FailedNodesMap{}

	return &schedulerapi.ExtenderFilterResult{
		Nodes:       nil,
		NodeNames:   args.NodeNames,
		FailedNodes: failedNodesMap,
	}, nil
}

func (e *Extender) Prioritize(args *schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	result := schedulerapi.HostPriorityList{}

	for _, node := range *args.NodeNames {
		result = append(result, schedulerapi.HostPriority{Host: node, Score: 1})
	}

	return &result, nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	extender1 := &Extender{}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		extender1.serveHTTP(w, req)
	})

	err := http.ListenAndServe(":8123", nil)
	if err != nil {
		glog.Fatal(err)
	}
}
