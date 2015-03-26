#!/bin/bash
#检查消息队列

#等待时间
wait_time=1

while true
do
    echo $(date '+%s') > /data/pid/$$

    ##要执行的程序

    sleep ${wait_time}
done