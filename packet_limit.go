package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func cvHumanReadableToPacketBytes(x string) (int64, error) {
	var mag int64 = 1
	if strings.HasSuffix(x, "GB") {
		x = strings.TrimSuffix(x, "GB")
		mag = 1 * 1024 * 1024 * 1024
	} else if strings.HasSuffix(x, "MB") {
		x = strings.TrimSuffix(x, "MB")
		mag = 1 * 1024 * 1024
	} else if strings.HasSuffix(x, "KB") {
		x = strings.TrimSuffix(x, "KB")
		mag = 1 * 1024
	}
	atoi, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0, err
	}
	return atoi * mag, nil

}
func main() {
	device := flag.String("d", "eth0", "监听的网卡")
	rxDefault := flag.String("rx", "1024GB", "进入的流量，GB/MB/KB")
	txDefault := flag.String("tx", "1024GB", "传出的流量，GB/MB/KB")
	cmdDefault := flag.String("e", "", "达到后执行的操作")
	times := flag.Int64("s", 60, "每隔多少秒执行一次")
	flag.Parse()
	log.Printf("程序已经运行,每隔 %d 秒检测网卡:%s是否到达警戒线。入口流量 : %s 出口流量 : %s\n", *times, *device, *rxDefault, *txDefault)
	rxValue, err := cvHumanReadableToPacketBytes(*rxDefault)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	txValue, err := cvHumanReadableToPacketBytes(*txDefault)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	ticker := time.NewTicker(time.Duration(*times) * time.Second)
	for {
		output, err := os.ReadFile("/proc/net/dev")
		if err != nil {
			log.Printf("读取文件错误: %v\n", err)
			return
		}
		sp := strings.Split(string(output), "\n")
		for _, spi := range sp {
			spTrim := strings.TrimSpace(spi)
			if strings.Contains(spTrim, *device) {
				rx, tx := calculate(strings.Split(spTrim, ":")[1])
				if rx == nil || tx == nil {
					log.Printf("计算错误\n")
					return
				}
				log.Printf("TX: %+v RX: %+v\n", tx, rx)
				rxBytes, err := strconv.ParseInt(rx.Bytes, 10, 64)
				if err != nil {
					return
				}
				txBytes, err := strconv.ParseInt(rx.Bytes, 10, 64)
				if err != nil {
					return
				}
				if rxBytes >= rxValue || txBytes >= txValue {
					log.Printf("预设值tx: %d rx: %d 当前值tx: %d rx: %d,执行命令 %s\n", txValue, rxValue, txBytes, rxBytes, *cmdDefault)
					opt, err := exec.Command("sh", "-c", *cmdDefault).Output()
					if err != nil {
						log.Printf("执行命令失败%v\n", err)
					} else {
						log.Printf("执行命令成功%s\n%s\n", *cmdDefault, string(opt))
					}
					return
				}
			}
		}
		<-ticker.C
	}

}
func calculate(str string) (*Packet, *Packet) {
	str = strings.TrimSpace(str)
	compile, err := regexp.Compile(`\s+`)
	if err != nil {
		return nil, nil
	}
	split := compile.Split(str, -1)
	return &Packet{
			Bytes:      split[0],
			Packets:    split[1],
			Errors:     split[2],
			Dropped:    split[3],
			FIFO:       split[4],
			Frame:      split[5],
			Compressed: split[6],
			Multicast:  split[7],
		}, &Packet{
			Bytes:      split[8],
			Packets:    split[9],
			Errors:     split[10],
			Dropped:    split[11],
			FIFO:       split[12],
			Frame:      split[13],
			Compressed: split[14],
			Multicast:  split[15],
		}
}

type Packet struct {
	Bytes      string
	Packets    string
	Errors     string
	Dropped    string
	FIFO       string
	Frame      string
	Compressed string
	Multicast  string
}
