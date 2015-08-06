package main

import (
	"bufio"
	log4go "code.google.com/p/log4go"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	//配置文件所在位置
	configPath = "/etc/xyDog.conf"
)

type processStruct struct {
	dir        string
	dirAndName string
	log        string
	num        int
}

var (
	logFile     string
	slowTime    int64
	pidDir      string
	processList map[string]processStruct
)

func main() {
	//创建异常处理
	defer func() {
		if err := recover(); err != nil {
			log4go.Error(fmt.Sprintf("panic error:%s", err.(string)))

			time.Sleep(10 * time.Millisecond)
		}
	}()

	//获取配置
	getConfig()

	//创建日志类
	logOption := log4go.NewFileLogWriter(logFile, false)
	log4go.AddFilter("file", log4go.FINE, logOption)

	for processName, processStruct := range processList {
		command := fmt.Sprintf("ps ax | grep -v 'grep' | grep '%s%s' | awk '{print $1}'", "/bin/sh ", processName)

		out, err := exec.Command("/bin/sh", "-c", command).Output()

		if err != nil {
			log4go.Error(fmt.Sprintf("命令执行失败:%s,错误代码:%s", command, err))
		}

		//程序的路径,检测shell文件状态
		filepath := fmt.Sprintf(processStruct.dirAndName)

		finfo, err := os.Stat(filepath)

		if err != nil {
			log4go.Error(fmt.Sprintf("无法找到shell文件:%s", filepath))
			continue
		}

		if finfo.IsDir() {
			log4go.Error(fmt.Sprintf("shell文件不存在，存在shell名称的文件夹:%s", filepath))
			continue
		}

		splitOut := strings.Fields(fmt.Sprintf("%s", out))

		//检测是否存在该进程，如果不存在则启动它，可能为首次启动～
		if len(splitOut) < processStruct.num {
			//根据缺少的进程数启动相应数目的进程
			subLen := processStruct.num - len(splitOut)

			for i := 0; i < subLen; i++ {
				log4go.Info(fmt.Sprintf("检测到进程不足，开始启动:%s", processName))

				startProcess(processName)
			}

		} else {
			//异常情况，进程比配置还多
			if len(splitOut) > processStruct.num {
				nowNum := len(splitOut)

				for _, pid := range splitOut {
					pid, err := strconv.Atoi(pid)

					if err != nil {
						log4go.Error(fmt.Sprintf("pid参数转换异常%s,程序为%s,错误代码:%s", pid, processName, err))
						continue
					}

					subLen := nowNum - processStruct.num

					if subLen > 0 {
						//获取进程结构体
						killProcessStruct, err := os.FindProcess(pid)

						//杀死进程
						err = killProcessStruct.Kill()

						if err != nil {
							log4go.Error(fmt.Sprintf("进程过多中，杀死进程失败%d,程序为:%s,错误代码:%s", pid, processName, err))
							continue
						} else {
							log4go.Info(fmt.Sprintf("进程过多中，杀死进程%d,程序为:%s", pid, processName))
						}

						nowNum -= 1
					} else {
						break
					}

				}
			}

			if len(splitOut) == 1 {
				//进程已存在,检测相关进程是否正常，如果没有则进程可能已经挂掉了，得启动它
				pid, err := strconv.Atoi(splitOut[0])

				if err != nil {
					log4go.Error(fmt.Sprintf("pid参数转换异常%s,程序为%s,错误代码:%s", pid, processName, err))
					continue
				}

				isNormal := startCheck(pid, processName)

				//不正常则需要重启进程
				if !isNormal {
					restartProcess(processName, pid)
				}
			} else if len(splitOut) > 1 {
				for _, pid := range splitOut {
					pid, err := strconv.Atoi(pid)

					if err != nil {
						log4go.Error(fmt.Sprintf("pid参数转换异常%s,程序为%s,错误代码:%s", pid, processName, err))
						continue
					}

					//首先检测进程是否运行正常
					isNormal := startCheck(pid, processName)

					//不正常则需要重启进程
					if !isNormal {
						restartProcess(processName, pid)
					}
				}
			}
		}

	}

	log4go.Info("看门狗正常运行")

	//为写入日志文件而停留一会儿
	time.Sleep(10 * time.Millisecond)
}

/**
 * 检测pid的文件是否存在并且时间符合要求
 */
func startCheck(pid int, processName string) bool {
	timeContent := ""

	//由于有时候shell脚本写文件时读取文件可能造成空的情况，因此重试多次。
	for i := 0; i < 3; i++ {
		content, err := ioutil.ReadFile(fmt.Sprintf(pidDir+"%d", pid))

		if err != nil {
			log4go.Info(fmt.Sprintf("pid文件不存在%d,程序为:%s,错误为:%s", pid, processName, err))
			return false
		}

		if len(content) != 0 {
			timeContent = strings.Trim(fmt.Sprintf("%s", content), "\n")

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	//文件如果存在的话需要检测时间是否符合要求
	fTime, err := strconv.ParseInt(timeContent, 10, 64)

	if err != nil {
		log4go.Info(fmt.Sprintf("文件内容无法转换,转换前为:%s,程序为:%s,错误为:%s", timeContent, processName, err))
		return false
	}

	//相差的时间
	subTime := slowTime - (time.Now().Unix() - fTime)

	if subTime < 0 {
		return false
	} else {
		return true
	}
}

/**
 * 通过程序名称启动进程
 */
func startProcess(processName string) bool {
	_, err := os.Stat("/usr/bin/nohup")
	if err != nil {
		log4go.Error(fmt.Sprintf("nohup命令无法找到:", err))
		return false
	}

	_, err = os.Stat("/bin/sh")
	if err != nil {
		log4go.Error(fmt.Sprintf("sh命令无法找到:", err))
		return false
	}

	//打开文件
	f, err := os.OpenFile(processList[processName].log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		log4go.Error(fmt.Sprintf("打开文件异常:%s", err))
	}

	cmd := exec.Command("/bin/sh", processName)

	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Dir = processList[processName].dir

	err = cmd.Start()

	if err != nil {
		log4go.Error(fmt.Sprintf("运行失败,错误代码为:%s", err))
	}

	f.Close()

	log4go.Info(fmt.Sprintf("启动进程:%s", processName))

	return true
}

/**
 * 重新启动程序
 * 为了兼容同时运行多个程序的要求，在这里通过pid来杀死进程
 */
func restartProcess(processName string, pid int) bool {
	//获取进程结构体
	processStruct, err := os.FindProcess(pid)

	if err != nil {
		log4go.Error(fmt.Sprintf("无法获取相关的进程%d,程序为:%s,错误代码:%s", pid, processName, err))
		return false
	}

	//杀死卡死的进程
	err = processStruct.Kill()

	if err != nil {
		log4go.Error(fmt.Sprintf("杀死进程失败%d,程序为:%s,错误代码:%s", pid, processName, err))
		return false
	} else {
		log4go.Info(fmt.Sprintf("杀死进程%d,程序为:%s", pid, processName))
	}

	//重新启动进程
	startProcess(processName)

	return true
}

/**
 * 获取配置文件
 */
func getConfig() {
	configMap := make(map[string]string)

	f, err := os.Open(configPath)

	//获取配置失败的话直接退出程序
	if err != nil {
		panic(fmt.Sprintf("无法读取配置文件，程序强制退出:%s", err))
	}

	buf := bufio.NewReader(f)

	for {
		l, err := buf.ReadString('\n')

		line := strings.TrimSpace(l)

		if err != nil {
			if err != io.EOF {
				panic(fmt.Sprintf("解析配置文件异常，程序强制退出:%s", err))
			}

			if len(line) == 0 {
				break
			}
		}

		switch {
		case len(line) == 0:
			break
		case line[0] == '#':
			break
		default:
			i := strings.IndexAny(line, "=")
			configMap[strings.TrimSpace(line[0:i])] = strings.TrimSpace(line[i+1:])
		}
	}

	//日志路径
	logFile = configMap["logFile"]

	if len(logFile) == 0 {
		panic(fmt.Sprintf("缺少logFile配置，程序强制退出"))
	}

	//可用的间隔时间
	slowTime, err = strconv.ParseInt(configMap["slowTime"], 10, 64)

	if err != nil || slowTime <= 0 {
		panic(fmt.Sprintf("缺少slowTime配置或者slowTime小于等于0，程序强制退出"))
	}

	//程序所写入的文件目录
	pidDir = configMap["pidDir"]

	if len(pidDir) == 0 {
		panic(fmt.Sprintf("缺少pidDir配置，程序强制退出"))
	}

	processList = make(map[string]processStruct)

	for key, value := range configMap {
		num := strings.LastIndex(key, ".sh")
		if num != -1 {
			numAndLog := strings.Split(value, ">>")

			intValue, err := strconv.ParseInt(strings.TrimSpace(numAndLog[0]), 10, 0)

			if err != nil {
				panic(fmt.Sprintf("shell配置后必须为数目，程序强制退出:%s", err))
			}

			nameLoc := strings.LastIndex(key, "/")

			dir := strings.TrimSpace(key[0:nameLoc])

			processName := strings.TrimSpace(key[nameLoc+1:])

			processList[processName] = processStruct{dir, key, strings.TrimSpace(numAndLog[1]), int(intValue)}
		}
	}
}
