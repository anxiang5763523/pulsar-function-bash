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

package runtime

import (
	"fmt"
	"github.com/anxiang5763523/pulsar-function-bash/conf"
)

// This is the config passed to the Golang Instance. Contains all the information
// passed to run functions
type instanceConf struct {
	instanceID                  int
	funcID                      string
	funcVersion                 string
	funcPath                    string
	funcDetails                 FunctionDetails
	pulsarServiceURL            string
}

func newInstanceConf() *instanceConf {
	config := &conf.Conf{}
	cfg := config.GetConf()
	if cfg == nil {
		panic("config file is nil.")
	}

	instanceConf := &instanceConf{
		instanceID:                  cfg.InstanceID,
		funcID:                      cfg.FuncID,
		funcVersion:                 cfg.FuncVersion,
		funcPath:                    cfg.FuncPath,
		pulsarServiceURL:            cfg.PulsarServiceURL,
		funcDetails: FunctionDetails{
			Tenant:               cfg.Tenant,
			Namespace:            cfg.NameSpace,
			Name:                 cfg.Name,
			LogTopic:             cfg.LogTopic,
			ProcessingGuarantees: ProcessingGuarantees(cfg.ProcessingGuarantees),
			AutoAck:              cfg.AutoACK,
			Parallelism:          cfg.Parallelism,
			Source: &SourceSpec{
				SubscriptionType: SubscriptionType(cfg.SubscriptionType),
				InputSpecs: map[string]*ConsumerSpec{
					cfg.SourceSpecTopic: {
						IsRegexPattern: cfg.IsRegexPatternSubscription,
						ReceiverQueueSize: &ConsumerSpec_ReceiverQueueSize{
							Value: cfg.ReceiverQueueSize,
						},
					},
				},
				TimeoutMs:            cfg.TimeoutMs,
				SubscriptionName:     cfg.SubscriptionName,
				CleanupSubscription:  cfg.CleanupSubscription,
				SubscriptionPosition: SubscriptionPosition(cfg.SubscriptionPosition),
			},
			Sink: &SinkSpec{
				Topic:      cfg.SinkSpecTopic,
			},
			RetryDetails: &RetryDetails{
				MaxMessageRetries: cfg.MaxMessageRetries,
				DeadLetterTopic:   cfg.DeadLetterTopic,
			},
			UserConfig: cfg.UserConfig,
		},
	}
	return instanceConf
}

func (ic *instanceConf) getInstanceName() string {
	return "" + fmt.Sprintf("%d", ic.instanceID)
}
