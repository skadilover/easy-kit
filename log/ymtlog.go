/*
@author: xuchengxuan(bigpyer@126.com)
@brief: logger module
*/
package log

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type LogConfig struct {
	Path  string
	Name  string
	Mode  int
	Level int
}

var logConfig *LogConfig

func Initialize_Base_Logger(path, name string, mode, level int) {
	logConfig = &LogConfig{
		Path:  path,
		Name:  name,
		Mode:  mode,
		Level: level,
	}
	_log = NewLogger(logConfig.Path, logConfig.Name)
	_log.logLevel = logConfig.Level
}

func Initialize_Base_Logger_with_config(c *LogConfig) {
	logConfig = c
	_log = NewLogger(logConfig.Path, logConfig.Name)
	_log.logLevel = logConfig.Level
}

/*
*INFO和ERROR接口必须有LogHeader参数,用于日志信息分析
*DEBUG接口没有LogHeader参数用于程序员编程调试使用
 */
const (
	DATEFORMAT        = "2006-01-02"
	DEFAULT_LOG_SCAN  = 300
	DEFAULT_LOG_LEVEL = DEBUG
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)
const (
	Info_str  = "INFO"
	Error_str = "ERROR"
	Debug_str = "DEBUG"
	Warn_str  = "WARN"
)
const (
	MOD_NORMAL = iota
	MOD_JSON
)

type LogObject struct {
	Timestamp int64                  `json:"timestamp"`
	Level     string                 `json:"level"`
	Logid     string                 `json:"logid"`
	Product   string                 `json:"product"`
	Module    string                 `json:"module"`
	Caller_ip string                 `json:"caller_ip"`
	Host_ip   string                 `json:"host_ip"`
	Msg       interface{}            `json:"msg"`
	Trace     map[string]interface{} `json:"trace"`
	Tag       string                 `json:"tag"`
}

type LogHeader struct {
	LogId    string
	ReqId    string
	HostId   string
	CallerIp string
	HostIp   string
	Product  string
	Module   string
	Lat      string
	Lng      string
}

var _log *logger = nil

type logger struct {
	mu       *sync.RWMutex
	fileDir  string
	fileName string

	date time.Time

	logFile  *os.File
	lger     *log.Logger
	timeScan int64

	logChan  chan string
	objChan  chan *LogObject
	logLevel int
}

func NewLogger(dir string, name string) *logger {
	dailyLogger := &logger{
		mu:       new(sync.RWMutex),
		fileDir:  dir,
		fileName: name,
		logChan:  make(chan string, 1024),
		objChan:  make(chan *LogObject, 1024),
		logLevel: DEFAULT_LOG_LEVEL,
	}

	dailyLogger.initDailyLogger()

	return dailyLogger
}

func (l *logger) initDailyLogger() {
	l.date, _ = time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	l.mu.Lock()
	defer l.mu.Unlock()

	logFile := joinFilePath(l.fileDir, l.fileName)
	l.logFile, _ = os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if logConfig.Mode == MOD_JSON {
		l.lger = log.New(l.logFile, "", 0)
	} else {
		l.lger = log.New(l.logFile, "", log.LstdFlags|log.Lmicroseconds)
	}

	go l.writeLog()
	go l.monitorFile()
}

func (l *logger) isNeedRotate() bool {
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	if t.After(l.date) {
		return true
	}
	return false
}

func (l *logger) rotate() {
	logFile := joinFilePath(l.fileDir, l.fileName)
	originBakName := logFile + "." + l.date.Format(DATEFORMAT)
	logFileBak := originBakName
	for i := 0; ; i++ {
		if isExist(logFileBak) {
			logFileBak = originBakName + "_" + strconv.Itoa(i)
		} else {
			break
		}
	}
	if l.logFile != nil {
		l.logFile.Close()
	}
	err := os.Rename(logFile, logFileBak)
	if err != nil {
		l.lger.Printf("logger rename error: %v", err.Error())
	}

	l.logFile, _ = os.Create(logFile)
	if logConfig.Mode == MOD_JSON {
		l.lger = log.New(l.logFile, "", 0)
	} else {
		l.lger = log.New(l.logFile, "", log.LstdFlags|log.Lmicroseconds)
	}
}

func (l *logger) monitorFile() {
	defer func() {
		if err := recover(); err != nil {
			l.lger.Panic("logger's FileMonitor() catch panic: %v\n", err)
		}
	}()

	// check frequency
	logScan := DEFAULT_LOG_SCAN

	timer := time.NewTicker(time.Duration(logScan) * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			l.checkFile()
		}
	}
}

func (l *logger) checkFile() {
	if l.isNeedRotate() {
		l.mu.Lock()
		l.rotate()
		l.mu.Unlock()
		l.date, _ = time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	}
}

// passive to close filelogger
func (l *logger) Close() error {

	close(l.logChan)
	l.lger = nil

	return l.logFile.Close()
}

// Receive logStr from f's logChan and print logstr to file
func (f *logger) writeLog() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf(" writeLog catch panic: %v\n", err)
		}
	}()

	for {
		select {
		case str := <-f.logChan:
			f.outPut(str)
		case obj := <-f.objChan:
			f.outPutObj(obj)
		}
	}
}

// print log
func (l *logger) outPut(str string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	l.lger.Output(2, str)
}

func (l *logger) outPutObj(o *LogObject) {
	d, err := json.Marshal(o)
	if err != nil {
		fmt.Println("Output unmarshal error.", err)
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lger.Printf("%s", d)
}

//Tag log
func Tag(header LogHeader, tag string, msg interface{}) {
	_log.Tag(header, tag, msg)
}

//info log
func Info(header LogHeader, format string, v ...interface{}) {
	_log.Info(header, format, v...)
}
func InfoJson(header LogHeader, msg interface{}) {
	_log.InfoJson(header, msg)
}

func (l *logger) Tag(header LogHeader, tag string, msg interface{}) {
	_, file, line, _ := runtime.Caller(2) //calldepth=3
	go func() {
		if l.logLevel <= INFO {
			switch logConfig.Mode {
			case MOD_NORMAL:
				logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.HostId)
				l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[INFO] "+logHeader, msg)
			case MOD_JSON:
				o := &LogObject{
					Timestamp: time.Now().Unix(),
					Level:     Info_str,
					Logid:     header.LogId,
					Product:   header.Product,
					Module:    header.Module,
					Caller_ip: header.CallerIp,
					Host_ip:   header.HostIp,
					Msg:       msg,
					Trace:     make(map[string]interface{}),
					Tag:       tag,
				}
				o.Trace["File"] = shortFileName(file)
				o.Trace["Line"] = line
				l.objChan <- o
			}
		}
	}()
}

func (l *logger) InfoJson(header LogHeader, msg interface{}) {
	_, file, line, _ := runtime.Caller(2) //calldepth=3
	go func() {
		if l.logLevel <= INFO {
			switch logConfig.Mode {
			case MOD_NORMAL:
				logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.HostId)
				l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[INFO] "+logHeader, msg)
			case MOD_JSON:
				o := &LogObject{
					Timestamp: time.Now().Unix(),
					Level:     Info_str,
					Logid:     header.LogId,
					Product:   header.Product,
					Module:    header.Module,
					Caller_ip: header.CallerIp,
					Host_ip:   header.HostIp,
					Msg:       msg,
					Trace:     make(map[string]interface{}),
				}
				o.Trace["File"] = shortFileName(file)
				o.Trace["Line"] = line
				l.objChan <- o
			}
		}
	}()
}

// internal info log
func (l *logger) Info(header LogHeader, format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(3) //calldepth=3
	if l.logLevel <= INFO {
		switch logConfig.Mode {
		case MOD_NORMAL:
			logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.Module)
			l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[INFO] "+logHeader+format, v...)
		case MOD_JSON:
			o := &LogObject{
				Timestamp: time.Now().Unix(),
				Level:     Info_str,
				Logid:     header.LogId,
				Product:   header.Product,
				Module:    header.Module,
				Caller_ip: header.CallerIp,
				Host_ip:   header.HostIp,
				Msg:       fmt.Sprintf(format, v...),
				Trace:     make(map[string]interface{}),
			}
			o.Trace["File"] = shortFileName(file)
			o.Trace["Line"] = line
			l.objChan <- o
		}
	}
}

// debug log
func Debug(header LogHeader, format string, v ...interface{}) {
	_log.Debug(header, format, v...)
}

// internal debug log
func (l *logger) Debug(header LogHeader, format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2) //calldepth=3
	if l.logLevel <= DEBUG {
		switch logConfig.Mode {
		case MOD_NORMAL:
			logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.HostId)
			l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[DEBUG] "+logHeader+format, v...)
		case MOD_JSON:
			o := &LogObject{
				Timestamp: time.Now().Unix(),
				Level:     Debug_str,
				Logid:     header.LogId,
				Product:   header.Product,
				Module:    header.Module,
				Caller_ip: header.CallerIp,
				Host_ip:   header.HostIp,
				Msg:       fmt.Sprintf(format, v...),
				Trace:     make(map[string]interface{}),
			}
			o.Trace["File"] = shortFileName(file)
			o.Trace["Line"] = line
			l.objChan <- o
		}
	}
}

// warn log
func Warn(header LogHeader, format string, v ...interface{}) {
	_log.Warn(header, format, v...)
}

// internal warn log
func (l *logger) Warn(header LogHeader, format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2) //calldepth=3
	if l.logLevel <= WARN {
		switch logConfig.Mode {
		case MOD_NORMAL:
			logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.HostId)
			l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[WARN] "+logHeader+format, v...)
		case MOD_JSON:
			o := &LogObject{
				Timestamp: time.Now().Unix(),
				Level:     Warn_str,
				Logid:     header.LogId,
				Product:   header.Product,
				Module:    header.Module,
				Caller_ip: header.CallerIp,
				Host_ip:   header.HostIp,
				Msg:       fmt.Sprintf(format, v...),
				Trace:     make(map[string]interface{}),
			}
			o.Trace["File"] = shortFileName(file)
			o.Trace["Line"] = line
			l.objChan <- o
		}
	}
}

// error log
func Error(header LogHeader, format string, v ...interface{}) {
	_log.Error(header, format, v...)
}

// internal error log
func (l *logger) Error(header LogHeader, format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2) //calldepth=3
	if l.logLevel <= ERROR {
		switch logConfig.Mode {
		case MOD_NORMAL:
			logHeader := fmt.Sprintf("[%s][%s][%s] MSG:", header.LogId, header.ReqId, header.Module)
			l.logChan <- fmt.Sprintf("[%v:%v]", shortFileName(file), line) + fmt.Sprintf("[ERROR] "+logHeader+format, v...)
		case MOD_JSON:
			o := &LogObject{
				Timestamp: time.Now().Unix(),
				Level:     Error_str,
				Logid:     header.LogId,
				Product:   header.Product,
				Module:    header.Module,
				Caller_ip: header.CallerIp,
				Host_ip:   header.HostIp,
				Msg:       fmt.Sprintf(format, v...),
				Trace:     make(map[string]interface{}),
			}
			o.Trace["File"] = shortFileName(file)
			o.Trace["Line"] = line
			l.objChan <- o
		}
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func joinFilePath(path, file string) string {
	return filepath.Join(path, file)
}

func shortFileName(file string) string {
	return filepath.Base(file)
}
