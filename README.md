# pulsar-function-bash
Develop Bash Functions

用户创建bash脚本，并在bash脚本中按照要求写函数，sh脚本的入参和出参要求如下
- 第一个参数是输入的消息内容
- 函数处理成功时通过echo "xxxx" 输出函数处理结果
- 函数处理失败时通过echo "xxxx" 1>&2 输出错误信息

    #! /bin/bash
    
    function exclamation()
    {
        result=$1"!"
    }
    
    exclamation $1
    echo $result

Package Bash Functions

通过以下命令打包函数

    FROM pulsar-function-sh-runner:[tag]
    
    WORKDIR /pulsar
    
    COPY [your bash function script] /pulsar/function

Deploy Pulsar Functions

- 创建configmap
  创建configmap，data字段中填充Bash Runtime进程需要的配置文件信息(配置文件字段参考设计文档)

- 部署函数
  通过statefulset方式部署函数


