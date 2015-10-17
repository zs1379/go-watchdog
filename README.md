# go-watchdog
go写的看门狗，用于监控服务器上的各种定时脚本的运行状态

花了半天研究了下golang，写的第一个程序，还是蛮实用的。

2015-10-17 经过在线上三个月的运行，一切稳定，平均每台自动重启了约400次的死亡进程，并且近2个月未出现问题，因此更新至1.0版本～

####注意----以下这种情况请不要使用本程序：需要保证有且仅有一条进程在运行，如果多了会造成并发问题的，这是由于该程序为了保证进程一定处于运行中，因此牺牲进程的单一性。

- 2015-10-17更新：增加同台服务器可以部署同名shell的特性。（注意：更新该版时会出现进程重复启动的问题，需要手动结束旧版本启动的进程）
- 2015-08-06更新：修复线上使用过程中碰到进程减少未自动杀死、shell写文件的时候读取文件会导致并发读写文件的异常等文件。
- 2015-06-04更新：修复几个小bug，当shell脚本与用户程序同名的时候会出现进程多次启动，以及提升对应的脚本在ubuntu与centos的兼容性。
- 2015-03-27更新：部署时发现有时候会同时部署多个系统在一台机子上，旧机子又是非docket的，木有环境隔离，因此修改程序，兼容同时检测多个路径下的shell文件。

###面向：
服务器上经常放着nohup的各种程序，例如秒级处理队列，秒级处理事项等，为了方便一般都采shel直接nohup+sleep 1来实现，但是有时候会碰到进程假死，进程挂掉，服务器重启等各种奇葩的情况。

###注意事项：
- 请将配置文件xyDog.conf放在/etc/xyDog.conf
- 请将watchDog的运行权限配置到定时任务所使用用户有权限执行的程度，同时建立相关目录
- 需要监控的shell里请加入    echo $(date '+%s') > /data/pid/$$   同时写入的路径需要与配置中的pidDir一致哈

````
#配置文件
#日志文件存放位置及名称
logFile = /data/watchDog.log

#最大未执行时间，用于判断进程是否假死的最大时间，超过该时间未写入文件则判断为假死 秒
slowTime = 300

#进程时间保存位置
pidDir = /data/pid/

#相关脚本所处位置以及可以启动的数量，格式需为xxx.sh = 数目 >> 脚本位置，注意必须为.sh文件
/data/server/shell_center/shell/android_push_handle.sh = 2 >> /tmp/nohup.log
/data/server/shell_center/shell/register_handle.sh = 1 >> /tmp/nohup.log
/data/server/shell_center/shell/message_handle.sh = 1 >> /tmp/nohup.log
/data/server/shell_center/shell/notice_handle.sh = 1 >> /tmp/nohup.log
````
