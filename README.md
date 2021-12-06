# Develop Bash Functions

用户创建bash脚本，并在bash脚本中按照要求写函数，sh脚本的入参和出参要求如下

- 第一个参数是输入的消息内容
- 函数处理成功时通过echo "xxxx" 输出函数处理结果
- 函数处理失败时通过echo "xxxx" 1>&2 输出错误信息

```
#! /bin/bash

function exclamation()
{
    result=$1"!"
}

exclamation $1
echo $result
```

# Package Bash Functions

通过以下命令打包函数

```
FROM pulsar-function-sh-runner:[tag]

WORKDIR /pulsar

COPY [your bash function script] /pulsar/function
```

# Deploy Pulsar Functions

- 创建configmap

  创建configmap，data字段中填充Bash Runtime进程需要的配置文件信息(**配置文件字段参考设计文档**)

  ```
  apiVersion: v1
  data:
    conf.yaml: |
      #
      # Licensed to the Apache Software Foundation (ASF) under one
      # or more contributor license agreements.  See the NOTICE file
      # distributed with this work for additional information
      # regarding copyright ownership.  The ASF licenses this file
      # to you under the Apache License, Version 2.0 (the
      # "License"); you may not use this file except in compliance
      # with the License.  You may obtain a copy of the License at
      #
      #   http://www.apache.org/licenses/LICENSE-2.0
      #
      # Unless required by applicable law or agreed to in writing,
      # software distributed under the License is distributed on an
      # "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
      # KIND, either express or implied.  See the License for the
      # specific language governing permissions and limitations
      # under the License.
      #
      pulsarServiceURL: "pulsar://10.96.175.178:6650"
      funcID: "ad2fd3ed-ad30-4e9e-8477-3e459e6feb5b"
      funcVersion: "1.0.0"
      # function details config
      tenant: "apache"
      nameSpace: "pulsar"
      name: "exclamation"
      logTopic: "persistent://apache/pulsar/log-topic-2"
      processingGuarantees: 0
      autoAck: true
      parallelism: 3
      # source config
      subscriptionType: 0
      subscriptionName: ""
      subscriptionPosition: 1
      # source input specs
      sourceSpecsTopic: persistent://apache/pulsar/test-topic-2
      isRegexPatternSubscription: false
      receiverQueueSize: 10
      # sink specs config
      sinkSpecsTopic: persistent://apache/pulsar/test-topic-output-2
      # retryDetails config
      maxMessageRetries: 0
      deadLetterTopic: ""
      funcPath: /pulsar/function
  kind: ConfigMap
  metadata:
    name: exclamation
    namespace: pulsar
  ```

- 部署函数

​       通过statefulset方式部署函数

```
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: exclamation
  namespace: pulsar
spec:
  selector:
    matchLabels:
      app: exclamation
  serviceName: "exclamation"
  replicas: 1
  template:
    metadata:
      labels:
        app: exclamation
    spec:
      terminationGracePeriodSeconds: 100
      containers:
      - name: exclamation
        image: [package bash function image]
        command: ["/pulsar/bin/shRuntime"]
        args: ["-instance-conf-path", "/pulsar/conf/conf.yaml"]
        volumeMounts:
        - name: config-volume
          mountPath: /pulsar/conf
      volumes:
      - name: config-volume
        configMap:
          # Provide the name of the ConfigMap containing the files you want
          # to add to the container
          name: exclamation
```

