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

type ProcessingGuarantees int
type SubscriptionPosition int
type SubscriptionType int
type FunctionState int

const (
	ProcessingGuarantees_ATLEAST_ONCE     ProcessingGuarantees = 0 // [default value]
	ProcessingGuarantees_ATMOST_ONCE      ProcessingGuarantees = 1
	ProcessingGuarantees_EFFECTIVELY_ONCE ProcessingGuarantees = 2
)

var (
	SubscriptionType_value = map[string]int{
		"SHARED":     0,
		"FAILOVER":   1,
		"KEY_SHARED": 2,
	}
)

type RetryDetails struct {
	MaxMessageRetries int
	DeadLetterTopic   string 
}

type FunctionDetails struct {
	Tenant               string                        
	Namespace            string                        
	Name                 string
	LogTopic             string                        
	ProcessingGuarantees ProcessingGuarantees         
	UserConfig           string
	AutoAck              bool                         
	Parallelism          int
	Source               *SourceSpec                  
	Sink                 *SinkSpec
	RetryDetails         *RetryDetails
	SubscriptionPosition SubscriptionPosition 
}

type ConsumerSpec struct {
	IsRegexPattern     bool                            
	ReceiverQueueSize  *ConsumerSpec_ReceiverQueueSize
	PoolMessages       bool                            
}

type ProducerSpec struct {
	MaxPendingMessages                 int
	MaxPendingMessagesAcrossPartitions int
	BatchBuilder                       string      
}

type SourceSpec struct {
	SubscriptionType SubscriptionType
	InputSpecs map[string]*ConsumerSpec 
	TimeoutMs  uint64
	TopicsPattern string
	SubscriptionName             string              
	CleanupSubscription          bool                 
	SubscriptionPosition         SubscriptionPosition
}

type SinkSpec struct {
	Topic          string        
	ProducerSpec   *ProducerSpec
}

type ConsumerSpec_ReceiverQueueSize struct {
	Value int
}
