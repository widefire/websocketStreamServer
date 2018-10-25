package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LOG_LEVEL_DISABLE = -1
	LOG_LEVEL_TRACE   = 0
	LOG_LEVEL_WARN    = 1
	LOG_LEVEL_DEBUG   = 2
	LOG_LEVEL_INFO    = 3
	LOG_LEVEL_ERROR   = 4
	LOG_LEVEL_FATAL   = 5

	LogChanCacheSize = 10 //10 chan 来缓冲
	logCacheSize     = 1024
)

func init() {
	logInstance.console = true
	logInstance.outputSeted = false
	logInstance.buffer = new(bytes.Buffer)
	logInstance.chans = make(chan []byte, LogChanCacheSize)
	logInstance.flag_log = LOG_SHORT_FILE
	go logInstance.threadLog()
}

type logInfo struct {
	level    int
	flag_log int
	console  bool
	buffer   *bytes.Buffer
	chans    chan []byte
	io.Writer
	sync.Mutex             //mutex writer
	mutexBuf    sync.Mutex //mutex buffer
	outputSeted bool
}

var logInstance logInfo

const (
	LOG_NO_FILE    = 0x0
	LOG_LONG_FILE  = 0x1
	LOG_SHORT_FILE = 0x2
	LOG_TIME       = 0x4
)

func SetLogLevel(l int) {
	logInstance.level = l
}

func OutputInCmd(inCmd bool) {
	logInstance.console = inCmd
}

func SetFlags(flag int) {
	logInstance.flag_log = flag
}

func SetOutput(w io.Writer) {
	logInstance.Lock()
	defer logInstance.Unlock()
	logInstance.Writer = w
	logInstance.outputSeted = true
}

func LOGT(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_TRACE {
		logInstance.getLogAppend(LOG_LEVEL_TRACE, v)
	}
}

func LOGW(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_WARN {
		logInstance.getLogAppend(LOG_LEVEL_WARN, v)
	}
}

func LOGD(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_DEBUG {
		logInstance.getLogAppend(LOG_LEVEL_DEBUG, v)
	}
}

func LOGI(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_INFO {
		logInstance.getLogAppend(LOG_LEVEL_INFO, v)
	}
}

func LOGE(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_ERROR {
		logInstance.getLogAppend(LOG_LEVEL_ERROR, v)
	}
}

func LOGF(v ...interface{}) {
	if logInstance.level <= LOG_LEVEL_FATAL {
		logInstance.getLogAppend(LOG_LEVEL_FATAL, v)
		os.Exit(1)
	}
}

func (this *logInfo) getLogAppend(lvl int, v ...interface{}) (str string) {
	str = ""
	flag := 0
	//time
	flag = (this.flag_log & 0x4)
	if flag == 0x4 {
		t := time.Now()
		str += t.Format("[2006/01/02 15:04:05] ")

	}
	//lvl
	switch lvl {
	case LOG_LEVEL_TRACE:
		str += "[TRACE] "
	case LOG_LEVEL_WARN:
		str += "[WARN] "
	case LOG_LEVEL_DEBUG:
		str += "[DEBUG] "
	case LOG_LEVEL_INFO:
		str += "[INFO] "
	case LOG_LEVEL_ERROR:
		str += "[ERROR] "
	case LOG_LEVEL_FATAL:
		str += "[FATAL] "
	}
	//location
	flag = (this.flag_log & 0x3)
	if LOG_SHORT_FILE == flag || LOG_LONG_FILE == flag {
		_, file, line, ok := runtime.Caller(2)
		if false == ok {
			str += "???:0 "
		} else {
			if LOG_LONG_FILE == flag {
				str = file + " " + strconv.Itoa(line)
			} else if LOG_SHORT_FILE == flag {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				str += short + ":" + strconv.Itoa(line) + " "
			}
		}
	}
	strbrackets := fmt.Sprint(v)
	if len(strbrackets) > 0 {
		strbrackets = strings.TrimLeft(strbrackets, "[")
		strbrackets = strings.TrimRight(strbrackets, "]")
		str += strbrackets
	}
	str += "\r\n"
	if this.console {
		fmt.Print(str)
	}
	//	this.Lock()
	//	defer this.Unlock()
	//	if this.Writer != nil {
	//		this.Write([]byte(str))
	//	}
	this.chans <- []byte(str)
	return
}

func (this *logInfo) threadLog() {
	go this.threadFlush()
	for {
		select {
		case data := <-this.chans:
			if len(data) > 0 {
				this.mutexBuf.Lock()
				this.buffer.Write(data)
				if logCacheSize < this.buffer.Len() {
					dataLog := this.buffer.Bytes()
					this.buffer.Reset()
					this.mutexBuf.Unlock()
					this.writeLog(dataLog)
					continue
				}
				this.mutexBuf.Unlock()
			}
		}
	}
}

func (this *logInfo) threadFlush() {
	for {
		select {
		case <-time.After(time.Minute * 5): //定时flush一次
			this.flush()
		}
	}
}

func (this *logInfo) flush() {
	this.mutexBuf.Lock()
	defer this.mutexBuf.Unlock()
	if this.buffer.Len() > 0 {
		dataLog := this.buffer.Bytes()
		this.buffer.Reset()
		go this.writeLog(dataLog)
	}
}

func (this *logInfo) writeLog(dataLog []byte) {
	this.Lock()
	defer this.Unlock()
	if this.outputSeted {
		this.Write(dataLog)
	}
}
