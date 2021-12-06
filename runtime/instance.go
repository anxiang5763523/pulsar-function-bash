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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"

	log "github.com/anxiang5763523/pulsar-function-bash/logutil"
)

type goInstance struct {
	function          function
	context           *FunctionContext
	producer          pulsar.Producer
	consumers         map[string]pulsar.Consumer
	client            pulsar.Client
	properties        map[string]string
}

// newGoInstance init goInstance and init function context
func newGoInstance() *goInstance {
	goInstance := &goInstance{
		context:   NewFuncContext(),
		consumers: make(map[string]pulsar.Consumer),
	}

	goInstance.context.outputMessage = func(topic string) pulsar.Producer {
		producer, err := goInstance.getProducer(topic)
		if err != nil {
			log.Fatal(err)
		}
		return producer
	}

	goInstance.properties = make(map[string]string)
	return goInstance
}

func (gi *goInstance) process(channel chan pulsar.ConsumerMessage,wg *sync.WaitGroup) error {
	for {
		select {
		case cm := <-channel:
			msgInput := cm.Message
			atMostOnce := gi.context.instanceConf.funcDetails.ProcessingGuarantees == ProcessingGuarantees_ATMOST_ONCE
			atLeastOnce := gi.context.instanceConf.funcDetails.ProcessingGuarantees == ProcessingGuarantees_ATLEAST_ONCE
			autoAck := gi.context.instanceConf.funcDetails.AutoAck
			if autoAck && atMostOnce {
				gi.ackInputMessage(msgInput)
			}
			gi.addLogTopicHandler()


			output, err := gi.handlerMsg(msgInput)
			if err != nil {
				log.Errorf("handler message error:%v", err)
				if autoAck && atLeastOnce {
					gi.nackInputMessage(msgInput)
				}
				continue
			}
			gi.processResult(msgInput, output)
		}
	}
	wg.Done()
	return nil
}
func (gi *goInstance) startFunction(function function) error {
	gi.function = function

	err := gi.setupClient()
	if err != nil {
		log.Errorf("setup client failed, error is:%v", err)
		return err
	}
	err = gi.setupProducer()
	if err != nil {
		log.Errorf("setup producer failed, error is:%v", err)
		return err
	}
	channel, err := gi.setupConsumer()
	if err != nil {
		log.Errorf("setup consumer failed, error is:%v", err)
		return err
	}
	err = gi.setupLogHandler()
	if err != nil {
		log.Errorf("setup log appender failed, error is:%v", err)
		return err
	}

	if gi.context.instanceConf.funcDetails.Parallelism > 1 &&
		int(gi.context.instanceConf.funcDetails.Source.SubscriptionType) != SubscriptionType_value["SHARED"] {
		err = fmt.Errorf("Parallelism bigger than 1 but SubscriptionType is not shared")
		log.Errorf("%s",err.Error())
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(gi.context.instanceConf.funcDetails.Parallelism)
	for index := 0; index < gi.context.instanceConf.funcDetails.Parallelism; index++ {
		go gi.process(channel,&wg)
	}
	wg.Wait()
	gi.closeLogTopic()
	gi.close()
	return nil
}

func (gi *goInstance) setupClient() error {
	client, err := pulsar.NewClient(pulsar.ClientOptions{

		URL: gi.context.instanceConf.pulsarServiceURL,
	})
	if err != nil {
		log.Errorf("create client error:%v", err)
		return err
	}
	gi.client = client
	return nil
}

func (gi *goInstance) setupProducer() error {
	if gi.context.instanceConf.funcDetails.Sink.Topic != "" {
		log.Debugf("Setting up producer for topic %s", gi.context.instanceConf.funcDetails.Sink.Topic)
		producer, err := gi.getProducer(gi.context.instanceConf.funcDetails.Sink.Topic)
		if err != nil {
			log.Fatal(err)
		}

		gi.producer = producer
		return nil
	}

	return nil
}

func (gi *goInstance) getProducer(topicName string) (pulsar.Producer, error) {
	properties := getProperties(getDefaultSubscriptionName(
		gi.context.instanceConf.funcDetails.Tenant,
		gi.context.instanceConf.funcDetails.Namespace,
		gi.context.instanceConf.funcDetails.Name), gi.context.instanceConf.instanceID)

	batchBuilderType := pulsar.DefaultBatchBuilder

	if gi.context.instanceConf.funcDetails.Sink.ProducerSpec != nil {
		batchBuilder := gi.context.instanceConf.funcDetails.Sink.ProducerSpec.BatchBuilder
		if batchBuilder != "" {
			if batchBuilder == "KEY_BASED" {
				batchBuilderType = pulsar.KeyBasedBatchBuilder
			}
		}
	}

	producer, err := gi.client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   topicName,
		Properties:              properties,
		CompressionType:         pulsar.LZ4,
		BatchingMaxPublishDelay: time.Millisecond * 10,
		BatcherBuilderType:      batchBuilderType,
	})
	if err != nil {
		log.Errorf("create producer error:%s", err.Error())
		return nil, err
	}

	return producer, err
}

func (gi *goInstance) setupConsumer() (chan pulsar.ConsumerMessage, error) {
	subscriptionType := pulsar.Shared
	if int(gi.context.instanceConf.funcDetails.Source.SubscriptionType) == SubscriptionType_value["FAILOVER"] {
		subscriptionType = pulsar.Failover
	}

	funcDetails := gi.context.instanceConf.funcDetails
	subscriptionName := funcDetails.Tenant + "/" + funcDetails.Namespace + "/" + funcDetails.Name
	if funcDetails.Source != nil && funcDetails.Source.SubscriptionName != "" {
		subscriptionName = funcDetails.Source.SubscriptionName
	}

	properties := getProperties(getDefaultSubscriptionName(
		funcDetails.Tenant,
		funcDetails.Namespace,
		funcDetails.Name), gi.context.instanceConf.instanceID)

	channel := make(chan pulsar.ConsumerMessage)

	var (
		consumer  pulsar.Consumer
		topicName *TopicName
		err       error
	)

	for topic, consumerConf := range funcDetails.Source.InputSpecs {
		topicName, err = ParseTopicName(topic)
		if err != nil {
			return nil, err
		}

		log.Debugf("Setting up consumer for topic: %s with subscription name: %s", topicName.Name, subscriptionName)
		if consumerConf.ReceiverQueueSize != nil {
			if consumerConf.IsRegexPattern {
				consumer, err = gi.client.Subscribe(pulsar.ConsumerOptions{
					TopicsPattern:     topicName.Name,
					ReceiverQueueSize: int(consumerConf.ReceiverQueueSize.Value),
					SubscriptionName:  subscriptionName,
					Properties:        properties,
					Type:              subscriptionType,
					MessageChannel:    channel,
				})
			} else {
				consumer, err = gi.client.Subscribe(pulsar.ConsumerOptions{
					Topic:             topicName.Name,
					SubscriptionName:  subscriptionName,
					Properties:        properties,
					Type:              subscriptionType,
					ReceiverQueueSize: int(consumerConf.ReceiverQueueSize.Value),
					MessageChannel:    channel,
				})
			}
		} else {
			if consumerConf.IsRegexPattern {
				consumer, err = gi.client.Subscribe(pulsar.ConsumerOptions{
					TopicsPattern:    topicName.Name,
					SubscriptionName: subscriptionName,
					Properties:       properties,
					Type:             subscriptionType,
					MessageChannel:   channel,
				})
			} else {
				consumer, err = gi.client.Subscribe(pulsar.ConsumerOptions{
					Topic:            topicName.Name,
					SubscriptionName: subscriptionName,
					Properties:       properties,
					Type:             subscriptionType,
					MessageChannel:   channel,
				})

			}
		}

		if err != nil {
			log.Errorf("create consumer error:%s", err.Error())
			return nil, err
		}
		gi.consumers[topicName.Name] = consumer
	}
	return channel, nil
}

func (gi *goInstance) handlerMsg(input pulsar.Message) (output []byte, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gi.context.SetCurrentRecord(input)

	ctx = NewContext(ctx, gi.context)
	msgInput := input.Payload()
	return gi.function.process(ctx, msgInput)
}

func (gi *goInstance) processResult(msgInput pulsar.Message, output []byte) {
	atLeastOnce := gi.context.instanceConf.funcDetails.ProcessingGuarantees == ProcessingGuarantees_ATLEAST_ONCE
	atMostOnce := gi.context.instanceConf.funcDetails.ProcessingGuarantees == ProcessingGuarantees_ATMOST_ONCE
	autoAck := gi.context.instanceConf.funcDetails.AutoAck
	if output != nil && gi.context.instanceConf.funcDetails.Sink.Topic != "" {
		asyncMsg := pulsar.ProducerMessage{
			Payload: output,
		}
		// Attempt to send the message and handle the response
		gi.producer.SendAsync(context.Background(), &asyncMsg, func(messageID pulsar.MessageID,
			message *pulsar.ProducerMessage, err error) {
			if err != nil {
				if autoAck && atLeastOnce {
					gi.nackInputMessage(msgInput)
				}
				log.Fatal(err)
			} else {
				if autoAck && !atMostOnce {
					gi.ackInputMessage(msgInput)
				}
			}
		})
	} else if autoAck && !atMostOnce {
		gi.ackInputMessage(msgInput)
	}
}

// ackInputMessage doesn't produce any result, or the user doesn't want the result.
func (gi *goInstance) ackInputMessage(inputMessage pulsar.Message) {
	log.Debugf("ack input message topic name is: %s", inputMessage.Topic())
	gi.consumers[inputMessage.Topic()].Ack(inputMessage)
}

func (gi *goInstance) nackInputMessage(inputMessage pulsar.Message) {
	gi.consumers[inputMessage.Topic()].Nack(inputMessage)
}

func (gi *goInstance) setupLogHandler() error {
	if gi.context.instanceConf.funcDetails.LogTopic != "" {
		gi.context.logAppender = NewLogAppender(
			gi.client, //pulsar client
			gi.context.instanceConf.funcDetails.LogTopic, //log topic
			getDefaultSubscriptionName(gi.context.instanceConf.funcDetails.Tenant, //fqn
				gi.context.instanceConf.funcDetails.Namespace,
				gi.context.instanceConf.funcDetails.Name),
		)
		return gi.context.logAppender.Start()
	}
	return nil
}

func (gi *goInstance) addLogTopicHandler() {
	// Clear StrEntry regardless gi.context.logAppender is set or not
	defer func() {
		log.StrEntry = nil
	}()

	if gi.context.logAppender == nil {
		log.Error("the logAppender is nil, if you want to use it, please specify `--log-topic` at startup.")
		return
	}

	for _, logByte := range log.StrEntry {
		gi.context.logAppender.Append([]byte(logByte))
	}
}

func (gi *goInstance) closeLogTopic() {
	log.Info("closing log topic...")
	if gi.context.logAppender == nil {
		return
	}
	gi.context.logAppender.Stop()
	gi.context.logAppender = nil
}

func (gi *goInstance) close() {
	log.Info("closing go instance...")
	if gi.producer != nil {
		gi.producer.Close()
	}
	if gi.consumers != nil {
		for _, consumer := range gi.consumers {
			consumer.Close()
		}
	}
	if gi.client != nil {
		gi.client.Close()
	}
}

