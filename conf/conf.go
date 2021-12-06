//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//

package conf

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	log "pulsar-function-bash/logutil"
)

const ConfigPath = "conf/conf.yaml"

type Conf struct {
	PulsarServiceURL string        `json:"pulsarServiceURL" yaml:"pulsarServiceURL"`
	InstanceID       int           `json:"instanceID" yaml:"instanceID"`
	FuncID           string        `json:"funcID" yaml:"funcID"`
	FuncVersion      string        `json:"funcVersion" yaml:"funcVersion"`
	FuncPath         string        `json:"funcPath" yaml:"funcPath"`
	// function details config
	Tenant               string `json:"tenant" yaml:"tenant"`
	NameSpace            string `json:"nameSpace" yaml:"nameSpace"`
	Name                 string `json:"name" yaml:"name"`
	LogTopic             string `json:"logTopic" yaml:"logTopic"`
	ProcessingGuarantees int  `json:"processingGuarantees" yaml:"processingGuarantees"`
	AutoACK              bool   `json:"autoAck" yaml:"autoAck"`
	Parallelism          int  `json:"parallelism" yaml:"parallelism"`
	//source config
	SubscriptionType     int  `json:"subscriptionType" yaml:"subscriptionType"`
	TimeoutMs            uint64 `json:"timeoutMs" yaml:"timeoutMs"`
	SubscriptionName     string `json:"subscriptionName" yaml:"subscriptionName"`
	CleanupSubscription  bool   `json:"cleanupSubscription"  yaml:"cleanupSubscription"`
	SubscriptionPosition int  `json:"subscriptionPosition" yaml:"subscriptionPosition"`
	//source input specs
	SourceSpecTopic            string `json:"sourceSpecsTopic" yaml:"sourceSpecsTopic"`
	IsRegexPatternSubscription bool   `json:"isRegexPatternSubscription" yaml:"isRegexPatternSubscription"`
	ReceiverQueueSize          int  `json:"receiverQueueSize" yaml:"receiverQueueSize"`
	//sink spec config
	SinkSpecTopic  string `json:"sinkSpecsTopic" yaml:"sinkSpecsTopic"`
	//retryDetails config
	MaxMessageRetries           int  `json:"maxMessageRetries" yaml:"maxMessageRetries"`
	DeadLetterTopic             string `json:"deadLetterTopic" yaml:"deadLetterTopic"`
	UserConfig                  string `json:"userConfig" yaml:"userConfig"`
}

var (
	help         bool
	confFilePath string
)

func (c *Conf) GetConf() *Conf {
	flag.Parse()

	if help {
		flag.Usage()
	}

	if confFilePath != "" {
		yamlFile, err := ioutil.ReadFile(confFilePath)
		if err == nil {
			err = yaml.Unmarshal(yamlFile, c)
			if err != nil {
				log.Errorf("unmarshal yaml file error:%s", err.Error())
				return nil
			}
		} else if os.IsNotExist(err) {
			log.Errorf("conf file is not exist, err:%s", err.Error())
			return nil
		} else if !os.IsNotExist(err) {
			log.Errorf("load conf file failed, err:%s", err.Error())
			return nil
		}
	}

	return c
}

func init() {
	var defaultPath string
	if err := os.Chdir("../"); err == nil {
		defaultPath = ConfigPath
	}
	log.Infof("The default config file path is: %s", defaultPath)

	flag.BoolVar(&help, "help", false, "print help cmd")
	flag.StringVar(&confFilePath, "instance-conf-path", defaultPath, "config conf.yml filepath")
}
