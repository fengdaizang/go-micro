package log

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"tanghu.com/go-micro/common/util"
)

// Options defines the log package options.
type Options struct {
	ServiceName string
}

// loggerParams defines the log package params
type loggerParams struct {
	level     logrus.Level
	formatter *logrus.TextFormatter
	fileFlag  bool
	writer    *rotatelogs.RotateLogs
	init      bool
}

// moduleFormatter for override Format function
type moduleFormatter struct {
	module string
	*logrus.TextFormatter
}

const (
	defaultLogLevel    = "ERROR"
	defaultTimeFormat  = "2006-01-02 15:04:05.000 MST"
	defaultServiceName = "go-micro"

	moduleLogKey  = "procMsg"
	sysIDLogKey   = "sysId"
	sysNameLogKey = "sysName"
	svcIDLogKey   = "svcId"
	isCRLogKey    = "isCR"
	warnMsgLogKey = "warnMsg"
	errMsgLogKey  = "errMsg"

	logrusErrorKey = "error"
)

var (
	loggerCache = map[string]*logrus.Logger{}

	serviceName      string
	theLoggerParams  loggerParams
	loggerParamsOnce sync.Once
)

// Init a logger
func Init(opts *Options) error {
	var err error

	serviceName = opts.ServiceName
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	loggerParamsOnce.Do(func() {
		err = initLoggerParams()
	})

	for module, logger := range loggerCache {
		moduleFormatter := wrapFormatter(module, theLoggerParams.formatter)

		logger.Level = theLoggerParams.level
		logger.Formatter = moduleFormatter
		logger.ReportCaller = true

		if theLoggerParams.fileFlag {
			logger.AddHook(newLfsHook(theLoggerParams.writer, moduleFormatter))
		}
	}

	return err
}

// MustGetLogger must get non-nil Logger, otherwise panic
func MustGetLogger(module string) *logrus.Logger {
	lg := getLogger(module)
	if lg == nil {
		panic("nil logger")
	}

	return lg
}

// GetLogger try to get a Logger
func GetLogger(module string) *logrus.Logger {
	return getLogger(module)
}

// getLogger get a logger by module name
func getLogger(module string) *logrus.Logger {
	lg, ok := loggerCache[module]
	if ok {
		return lg
	}

	var logger *logrus.Logger
	if !theLoggerParams.init {
		logger = logrus.New()
		logger.ReportCaller = true
	} else {
		moduleFormatter := wrapFormatter(module, theLoggerParams.formatter)

		logger = logrus.New()
		logger.Level = theLoggerParams.level
		logger.Formatter = moduleFormatter
		logger.ReportCaller = true

		if theLoggerParams.fileFlag {
			logger.AddHook(newLfsHook(theLoggerParams.writer, moduleFormatter))
		}
	}

	loggerCache[module] = logger

	return logger
}

// wrapFormatter get a moduleFormatter to replace logrus.TextFormatter
func wrapFormatter(module string, formatter *logrus.TextFormatter) *moduleFormatter {
	return &moduleFormatter{
		module:        module,
		TextFormatter: formatter,
	}
}

// initLoggerParams init logger params
func initLoggerParams() error {
	var (
		err        error
		timeFormat string
		formatter  *logrus.TextFormatter
		level      logrus.Level
		writer     *rotatelogs.RotateLogs
	)

	// timeFormat
	timeFormat = viper.GetString("logging.timeFormat")
	if timeFormat == "" {
		timeFormat = defaultTimeFormat
	}

	// formatter
	formatter = &logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: timeFormat,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timeStamp",
			logrus.FieldKeyMsg:  "message",
		},
	}

	// logLevel
	logLevel := viper.GetString("logging.level")
	if logLevel == "" {
		logLevel = defaultLogLevel
	}
	level, err = logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	// fileFlag
	fileFlag := viper.GetBool("logging.file.enable")

	// theLoggerParams
	theLoggerParams.level = level
	theLoggerParams.formatter = formatter
	theLoggerParams.fileFlag = fileFlag
	theLoggerParams.init = true

	if fileFlag {
		filePath := viper.GetString("logging.file.path")
		if filePath != "" {
			err = util.CheckFilePath(filePath)
			if err != nil {
				return err
			}
		}
		fileName := path.Join(filePath, serviceName)
		writer, err = rotatelogs.New(
			// 分割后的文件名称
			fileName+".%Y%m%d.log",

			// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
			rotatelogs.WithLinkName(fileName),

			// WithRotationTime设置日志分割的时间，这里设置为一天分割一次
			// WithRotationSize设置日志分割的大小
			rotatelogs.WithRotationTime(24*time.Hour),

			// WithMaxAge和WithRotationCount二者只能设置一个，
			// WithMaxAge设置文件清理前的最长保存时间，
			// WithRotationCount设置文件清理前最多保存的个数。
			rotatelogs.WithMaxAge(7*24*time.Hour),
		)
		if err != nil {
			return err
		}

		theLoggerParams.writer = writer
	}

	return nil
}

// newLfsHook new a lfsHook for split log file
func newLfsHook(writer *rotatelogs.RotateLogs, tf *moduleFormatter) logrus.Hook {
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, tf)

	return lfsHook
}

// Format override logrus.TextFormatter.format function
func (tf *moduleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Data[moduleLogKey] = fmt.Sprintf("%s->pid:%d", tf.module, os.Getegid())
	entry.Data[sysIDLogKey] = viper.GetString("sys.id")
	entry.Data[sysNameLogKey] = viper.GetString("sys.name")
	entry.Data[svcIDLogKey] = serviceName

	errHappened := entry.Level == logrus.ErrorLevel || entry.Level == logrus.FatalLevel
	if errHappened {
		entry.Data[isCRLogKey] = entry.Level == logrus.FatalLevel
	}

	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	prefixFieldClashes(data, tf.FieldMap, entry.HasCaller())
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	fixedKeys := make([]string, 0, 3+len(data))
	if !tf.DisableTimestamp {
		fixedKeys = append(fixedKeys, resolve(tf.FieldMap, logrus.FieldKeyTime))
	}
	fixedKeys = append(fixedKeys, resolve(tf.FieldMap, logrus.FieldKeyLevel))
	fixedKeys = append(fixedKeys, resolve(tf.FieldMap, logrus.FieldKeyMsg))

	var funcVal, fileVal string
	if entry.HasCaller() {
		if tf.CallerPrettyfier != nil {
			funcVal, fileVal = tf.CallerPrettyfier(entry.Caller)
		} else {
			funcVal = entry.Caller.Function
			fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		}
	}

	if errHappened {
		fixedKeys = append(fixedKeys, errMsgLogKey)
	}

	if entry.Level == logrus.WarnLevel {
		fixedKeys = append(fixedKeys, warnMsgLogKey)
	}

	fixedKeys = append(fixedKeys, keys...)

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := tf.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimeFormat
	}
	for _, key := range fixedKeys {
		var value interface{}
		switch {
		case key == resolve(tf.FieldMap, logrus.FieldKeyTime):
			value = entry.Time.Format(timestampFormat)
		case key == resolve(tf.FieldMap, logrus.FieldKeyLevel):
			value = entry.Level.String()
		case key == resolve(tf.FieldMap, logrus.FieldKeyMsg):
			if entry.Level >= logrus.InfoLevel {
				value = fmt.Sprintf("%s@%s:%s", fileVal, funcVal, entry.Message)
			} else {
				value = fmt.Sprintf("%s@%s", fileVal, funcVal)
			}
		case key == errMsgLogKey:
			var ok bool
			value, ok = data[logrusErrorKey]
			if ok {
				value = data[logrusErrorKey]
			}

			if entry.Message != "" {
				value = fmt.Sprintf("%v:%s", value, entry.Message)
			}
		case key == warnMsgLogKey:
			value = entry.Message
		case key == logrusErrorKey:
			continue
		default:
			value = data[key]
		}
		tf.appendKeyValue(b, key, value)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// resolve if there are custom fields use custom fields, if not, is the default fields
func resolve(f logrus.FieldMap, key string) string {
	for k, v := range f {
		if string(k) == key {
			return v
		}
	}

	return key
}

func prefixFieldClashes(data logrus.Fields, fieldMap logrus.FieldMap, reportCaller bool) {
	timeKey := resolve(fieldMap, logrus.FieldKeyTime)
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}

	msgKey := resolve(fieldMap, logrus.FieldKeyMsg)
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}

	levelKey := resolve(fieldMap, logrus.FieldKeyLevel)
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}

	logrusErrKey := resolve(fieldMap, logrus.FieldKeyLogrusError)
	if l, ok := data[logrusErrKey]; ok {
		data["fields."+logrusErrKey] = l
		delete(data, logrusErrKey)
	}

	// If reportCaller is not set, 'func' will not conflict.
	if reportCaller {
		funcKey := resolve(fieldMap, logrus.FieldKeyFunc)
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := resolve(fieldMap, logrus.FieldKeyFile)
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
}

func (tf *moduleFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(',')
	}
	b.WriteString(key)
	b.WriteByte('=')
	tf.appendValue(b, value)
}

func (tf *moduleFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	b.WriteString(stringVal)
}
