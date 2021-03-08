package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	mlog "github.com/micro/go-micro/util/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	defaultLogLevel   = "ERROR"
	defaultTimeFormat = "2006-01-02 15:04:05.000 MST"

	microLoggerName = "go.micro.internal"

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
	defaultLogOutput = os.Stderr
	theLoggerParams  struct {
		output    io.Writer
		level     logrus.Level
		formatter *logrus.TextFormatter
		once      sync.Once
		init      bool
	}
	loggerCache = map[string]*moduleLogger{}

	serviceName string
)

type loggerParams struct {
	output    io.Writer
	level     logrus.Level
	formatter logrus.Formatter
	once      sync.Once
}

// Options defines the log package options.
type Options struct {
	ServiceName string
}

type moduleLogger struct {
	module string
	logger *logrus.Logger
}

type microLogger struct {
	logger *logrus.Logger
}

type moduleFormatter struct {
	module string
	*logrus.TextFormatter
}

// Init initializes the logger.
func Init(opts *Options) error {
	serviceName = opts.ServiceName
	lgr := getLogger(microLoggerName)
	mlog.SetLogger(&microLogger{lgr})

	initLoggerParams()
	for _, lg := range loggerCache {
		lg.logger.SetLevel(theLoggerParams.level)
		lg.logger.Formatter = wrapFormatter(lg.module, theLoggerParams.formatter)
		lg.logger.Out = theLoggerParams.output
	}
	return nil
}

func initLoggerParams() error {
	if theLoggerParams.init {
		return nil
	}

	var err error
	theLoggerParams.once.Do(func() {
		timeFormat := viper.GetString("logging.timeFormat")
		if timeFormat == "" {
			timeFormat = defaultTimeFormat
		}
		logLevel := viper.GetString("logging.level")
		if logLevel == "" {
			logLevel = defaultLogLevel
		}

		var (
			level logrus.Level
			err   error
		)
		level, err = logrus.ParseLevel(logLevel)
		if err != nil {
			return
		}

		theLoggerParams.output = defaultLogOutput
		theLoggerParams.level = level
		theLoggerParams.formatter = &logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: timeFormat,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "timeStamp",
				logrus.FieldKeyMsg:  "message",
			},
		}
		theLoggerParams.init = true
	})

	return err
}

// MustGetLogger must get non-nil Logger, otherwise panic
func MustGetLogger(module string) *logrus.Logger {
	l := getLogger(module)
	if l == nil {
		panic("nil logger")
	}

	return l
}

// GetLogger try to get a Logger
func GetLogger(module string) *logrus.Logger {
	return getLogger(module)
}

func getLogger(module string) (logger *logrus.Logger) {
	lg, ok := loggerCache[module]
	if ok {
		return lg.logger
	}

	if !theLoggerParams.init {
		logger = logrus.New()
		logger.ReportCaller = true
	} else {
		logger = &logrus.Logger{
			Out:          theLoggerParams.output,
			Formatter:    wrapFormatter(module, theLoggerParams.formatter),
			Hooks:        make(logrus.LevelHooks),
			Level:        theLoggerParams.level,
			ReportCaller: true,
		}
	}

	loggerCache[module] = &moduleLogger{module, logger}
	return
}

func wrapFormatter(module string, formatter *logrus.TextFormatter) *moduleFormatter {
	return &moduleFormatter{
		module:        module,
		TextFormatter: formatter,
	}
}

func (tf *moduleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// callpath := 7
	// if tf.module == microLoggerName {
	//     callpath = 10
	// }

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

func formatCallpath(calldepth int) string {
	v := "???"
	if pc, _, _, ok := runtime.Caller(calldepth + 1); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			v = formatFuncName(f.Name())
		}
	}

	return v
}

func formatFuncName(f string) string {
	i := strings.LastIndex(f, "/")
	j := strings.Index(f[i+1:], ".")
	if j < 1 {
		return "???"
	}
	fun := f[i+j+2:]

	i = strings.LastIndex(fun, ".")
	return fun[i+1:]
}

// them differently.
func (l *microLogger) Log(v ...interface{}) {
	l.logger.Print(v...)
}

func (l *microLogger) Logf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}
