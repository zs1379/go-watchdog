# go-watchdog
go写的看门狗，用于监控服务器上的各种定时脚本的运行状态

花了半天研究了下golang，写的第一个程序，还是蛮实用的。

###面向：
服务器上经常放着nohup的各种程序，例如秒级处理队列，秒级处理事项等，为了方便一般都采shel直接nohup+sleep 1来实现，但是有时候会碰到进程假死，进程挂掉，服务器重启等各种奇葩的情况。

###注意事项：
- 请将配置文件xyDog.conf放在/etc/xyDog.conf
- 需要监控的shell里请加入    echo $(date '+%s') > /data/pid/$$   同时写入的路径需要与配置中的pidDir一致哈

````
#配置文件
#日志文件存放位置及名称
logFile = /data/watchDog.log

#用于判断进程是否假死的最大时间，超过该时间未写入文件则判断为假死，自动重启进程 秒
slowTime = 300

#进程是否卡死的进程文件存放位置
pidDir = /data/pid/

#脚本文件所在的位置
baseDir = /data/server/shell_center/shell/

#相关脚本所处位置以及可以启动的数量，格式需为xxx.sh = 数目，注意必须为.sh文件
android_push_handle.sh = 2
register_handle.sh = 1
message_handle.sh = 1
notice_handle.sh = 1
````
