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

//
// This file borrows some implementations from
// {@link https://github.com/aws/aws-lambda-go/blob/master/lambda/handler.go}
//  - errorHandler
//  - validateArguments
//  - validateReturns
//  - NewFunction
//  - Process
//

package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	log "pulsar-function-bash/logutil"
	conf "pulsar-function-bash/conf"
)

type function interface {
	process(ctx context.Context, input []byte) ([]byte, error)
}

type pulsarFunction struct {
	funcPath string
}

func (function pulsarFunction) process(ctx context.Context, input []byte) ([]byte, error) {
	c := exec.Command("sh",function.funcPath, string(input))
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		log.Errorf("process function error:[%s]\n", err.Error())
		return nil,err
	} else if err == nil && errStr != "" {
		err = fmt.Errorf("%s",errStr)
		return nil, err
	}

	return []byte(outStr), nil
}

func newFunction() (function,error) {
     config := &conf.Conf{}
     cfg := config.GetConf()
     if cfg == nil {
         return nil,fmt.Errorf("conf file is nil")
     }
     funcPath := fmt.Sprintf("%s/%s.sh",cfg.FuncPath,cfg.Name)
     log.Infof("funcPath is %s\n", funcPath)
     pulsarFunc := pulsarFunction{
		    funcPath: funcPath,
	        }
     return pulsarFunc, nil
}

func Start() {
	var err error
	var pulsarFunc function
	pulsarFunc, err = newFunction()
	if err != nil {
		log.Fatal(err)
		panic("start function failed, please check.")
	}
	goInstance := newGoInstance()
	err = goInstance.startFunction(pulsarFunc)
	if err != nil {
		log.Fatal(err)
		panic("start function failed, please check.")
	}
}

// GetUserConfMap provides a means to access the pulsar function's user config
// map before initializing the pulsar function
func GetUserConfMap() map[string]interface{} {
	return NewFuncContext().userConfigs
}

// GetUserConfValue provides access to a user configuration value before
// initializing the pulsar function
func GetUserConfValue(key string) interface{} {
	return NewFuncContext().userConfigs[key]
}
