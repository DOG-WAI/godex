package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm/log"
)

func init() {
	setDefaultFormat()
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level" json:"level"`
	File     string `yaml:"file" json:"file"`
	Rotate   bool   `yaml:"rotate" json:"rotate"`     // 是否启用日志轮转
	MaxDays  int    `yaml:"max-days" json:"max-days"` // 保留多少天的日志文件，0表示不删除
	TimeZone string `yaml:"timezone" json:"timezone"` // 时区，例如: "Asia/Shanghai", 空则使用本地时区
}

// Logger 全局日志实例
var Logger *logrus.Logger

// RotatingFileWriter 支持按日轮转的文件写入器
type RotatingFileWriter struct {
	filename    string
	currentFile *os.File
	currentDate string
	mutex       sync.Mutex
	maxDays     int
	location    *time.Location
}

// NewRotatingFileWriter 创建新的轮转文件写入器
func NewRotatingFileWriter(filename string, maxDays int, timeZone string) (*RotatingFileWriter, error) {
	// 解析时区
	var loc *time.Location
	var err error
	if timeZone == "" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(timeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %s: %v", timeZone, err)
		}
	}

	rfw := &RotatingFileWriter{
		filename: filename,
		maxDays:  maxDays,
		location: loc,
	}

	// 初始化当前文件
	if err := rfw.rotate(); err != nil {
		return nil, err
	}

	return rfw, nil
}

// Write 实现 io.Writer 接口
func (rfw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rfw.mutex.Lock()
	defer rfw.mutex.Unlock()

	// 检查是否需要轮转
	now := time.Now().In(rfw.location)
	currentDate := now.Format("2006-01-02")

	if currentDate != rfw.currentDate {
		if err := rfw.rotate(); err != nil {
			return 0, err
		}
	}

	if rfw.currentFile == nil {
		return 0, fmt.Errorf("no current log file")
	}

	return rfw.currentFile.Write(p)
}

// rotate 执行日志轮转
func (rfw *RotatingFileWriter) rotate() error {
	now := time.Now().In(rfw.location)
	newDate := now.Format("2006-01-02")

	// 如果日期没变，不需要轮转
	if newDate == rfw.currentDate && rfw.currentFile != nil {
		return nil
	}

	// 关闭当前文件
	if rfw.currentFile != nil {
		rfw.currentFile.Close()
	}

	// 生成新的文件名
	dir := filepath.Dir(rfw.filename)
	baseName := filepath.Base(rfw.filename)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	newFilename := filepath.Join(dir, fmt.Sprintf("%s-%s%s", nameWithoutExt, newDate, ext))

	// 确保目录存在
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	// 打开新文件
	file, err := os.OpenFile(newFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %v", newFilename, err)
	}

	rfw.currentFile = file
	rfw.currentDate = newDate

	// 清理过期日志文件
	if rfw.maxDays > 0 {
		go rfw.cleanupOldLogs()
	}

	return nil
}

// cleanupOldLogs 清理过期的日志文件
func (rfw *RotatingFileWriter) cleanupOldLogs() {
	defer func() {
		if r := recover(); r != nil {
			// 静默处理清理过程中的错误，避免影响主程序
		}
	}()

	dir := filepath.Dir(rfw.filename)
	baseName := filepath.Base(rfw.filename)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// 计算过期时间
	cutoffTime := time.Now().In(rfw.location).AddDate(0, 0, -rfw.maxDays)

	// 遍历目录查找过期文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		// 检查是否是我们的日志文件格式: name.YYYY-MM-DD.ext
		prefix := nameWithoutExt + "."
		suffix := ext

		if !strings.HasPrefix(fileName, prefix) || !strings.HasSuffix(fileName, suffix) {
			continue
		}

		// 提取日期部分
		dateStr := strings.TrimPrefix(strings.TrimSuffix(fileName, suffix), prefix)
		if len(dateStr) != 10 { // YYYY-MM-DD 长度为10
			continue
		}

		// 解析日期
		logDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// 如果日志文件过期，删除它
		if logDate.Before(cutoffTime) {
			fullPath := filepath.Join(dir, fileName)
			os.Remove(fullPath)
		}
	}
}

// Close 关闭文件写入器
func (rfw *RotatingFileWriter) Close() error {
	rfw.mutex.Lock()
	defer rfw.mutex.Unlock()

	if rfw.currentFile != nil {
		return rfw.currentFile.Close()
	}
	return nil
}

// getRealCaller 获取真正的调用者信息，跳过 logger 包中的函数
func getRealCaller() (string, int) {
	// 获取更深的调用栈
	pcs := make([]uintptr, 20)
	n := runtime.Callers(3, pcs) // 从第3层开始，跳过 getRealCaller 和 CallerPrettyfier
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// 跳过 logger 包中的函数和 logrus 包中的函数
		if !strings.Contains(frame.File, "/pkg/logger/") &&
			!strings.Contains(frame.File, "github.com/sirupsen/logrus") &&
			!strings.Contains(frame.Function, "github.com/sirupsen/logrus") {
			return frame.File, frame.Line
		}
		if !more {
			break
		}
	}
	return "", 0
}

// formatCaller 格式化调用者信息
func formatCaller(file string, line int) string {
	if file == "" {
		return ""
	}

	// 简化文件路径，只保留相对于项目根目录的路径
	filename := file
	if idx := strings.LastIndex(filename, "godex/"); idx != -1 {
		filename = filename[idx+len("godex/"):] // 保留项目根目录后的路径
	}

	// 从后往前取最多2个路径部分
	parts := strings.Split(filename, "/")
	if len(parts) > 2 {
		// 只保留最后2个部分（倒数第二个目录 + 文件名）
		filename = strings.Join(parts[len(parts)-2:], "/")
	}

	return fmt.Sprintf("%s:%d", filename, line)
}

// customTextFormatter 自定义文本格式化器，用于显示caller字段
type customTextFormatter struct {
	*logrus.TextFormatter
}

func (f *customTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 如果有caller字段，修改消息内容
	if caller, ok := entry.Data["caller"]; ok {
		// 创建一个修改过的entry
		modifiedEntry := *entry // 复制结构体
		modifiedEntry.Data = make(logrus.Fields)
		// 复制所有字段除了caller
		for k, v := range entry.Data {
			if k != "caller" {
				modifiedEntry.Data[k] = v
			}
		}
		// 修改消息，添加caller信息
		modifiedEntry.Message = fmt.Sprintf("[%s] %s", caller, entry.Message)
		return f.TextFormatter.Format(&modifiedEntry)
	}
	return f.TextFormatter.Format(entry)
}

// 终端输出格式（文本格式，易读）
var terminalFmt = &customTextFormatter{
	TextFormatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05-07:00",
		ForceColors:     true,
	},
}

// 文件输出格式（JSON格式，便于处理）
var fileFmt = &logrus.JSONFormatter{
	TimestampFormat: "2006-01-02T15:04:05.000-0700",
	FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime:  "time",
		logrus.FieldKeyLevel: "level",
		logrus.FieldKeyMsg:   "msg",
		logrus.FieldKeyFunc:  "caller",
	},
	CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		realFile, realLine := getRealCaller()
		return "", formatCaller(realFile, realLine)
	},
}

// DualFormatHook 双格式输出钩子
type DualFormatHook struct {
	terminalLogger *logrus.Logger
	fileLogger     *logrus.Logger
}

// Levels 返回支持的日志级别
func (hook *DualFormatHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 触发钩子，同时输出到终端和文件
func (hook *DualFormatHook) Fire(entry *logrus.Entry) error {
	// 输出到终端（文本格式）
	if hook.terminalLogger != nil {
		hook.terminalLogger.WithFields(entry.Data).Log(entry.Level, entry.Message)
	}

	// 输出到文件（JSON格式）
	if hook.fileLogger != nil {
		hook.fileLogger.WithFields(entry.Data).Log(entry.Level, entry.Message)
	}

	return nil
}

// setDefaultFormat 设置默认的日志格式（用于程序启动早期）
// 确保整个应用启动过程的日志格式一致
func setDefaultFormat() {
	defaultFmt := &customTextFormatter{
		TextFormatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05-07:00",
			ForceColors:     true,
		},
	}
	logrus.SetFormatter(defaultFmt)
	logrus.SetReportCaller(false) // 关闭默认的caller，我们使用自定义的
}

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 关闭默认的调用者信息记录，我们使用自定义的
	Logger.SetReportCaller(false)

	// 如果只有终端输出，使用文本格式
	if config.File == "" {
		Logger.SetFormatter(terminalFmt)
		Logger.SetOutput(os.Stdout)
	} else {
		// 如果有文件输出，创建双格式输出

		// 创建终端logger（文本格式）
		terminalLogger := logrus.New()
		terminalLogger.SetFormatter(terminalFmt)
		terminalLogger.SetLevel(level)
		terminalLogger.SetReportCaller(false)
		terminalLogger.SetOutput(os.Stdout)

		var fileOutput io.Writer

		// 根据是否启用轮转选择不同的文件输出方式
		if config.Rotate {
			// 使用轮转文件写入器
			rotatingWriter, err := NewRotatingFileWriter(config.File, config.MaxDays, config.TimeZone)
			if err != nil {
				return fmt.Errorf("failed to create rotating file writer: %v", err)
			}
			fileOutput = rotatingWriter
		} else {
			// 使用普通文件输出
			// 确保日志目录存在
			logDir := filepath.Dir(config.File)
			if logDir != "." && logDir != "" {
				if err := os.MkdirAll(logDir, 0755); err != nil {
					return fmt.Errorf("failed to create log directory: %v", err)
				}
			}

			// 打开日志文件
			logFile, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return fmt.Errorf("failed to open log file: %v", err)
			}
			fileOutput = logFile
		}

		// 创建文件logger（JSON格式）
		fileLogger := logrus.New()
		fileLogger.SetFormatter(fileFmt)
		fileLogger.SetLevel(level)
		fileLogger.SetReportCaller(false)
		fileLogger.SetOutput(fileOutput)

		// 主logger不直接输出，通过hook分发
		Logger.SetOutput(io.Discard)
		Logger.AddHook(&DualFormatHook{
			terminalLogger: terminalLogger,
			fileLogger:     fileLogger,
		})
	}

	// 替换logrus的默认logger
	logrus.SetOutput(Logger.Out)
	logrus.SetLevel(Logger.Level)
	logrus.SetFormatter(Logger.Formatter)
	logrus.SetReportCaller(false)

	if config.Rotate {
		Logger.Infof("🎉 Logger initialized successfully with daily rotation (max_days: %d, timezone: %s)", config.MaxDays, config.TimeZone)
	} else {
		Logger.Info("🎉 Logger initialized successfully")
	}
	return nil
}

// GetXormLogger 获取适配XORM的日志适配器
func GetXormLogger() log.Logger {
	return &XormLoggerAdapter{
		logger:  Logger,
		showSQL: false, // 默认不显示，由engine.ShowSQL()调用设置
	}
}

// XormLoggerAdapter XORM日志适配器
type XormLoggerAdapter struct {
	logger  *logrus.Logger
	level   log.LogLevel
	showSQL bool
}

// Level 获取日志级别
func (x *XormLoggerAdapter) Level() log.LogLevel {
	return x.level
}

// SetLevel 设置日志级别
func (x *XormLoggerAdapter) SetLevel(l log.LogLevel) {
	x.level = l
}

// ShowSQL 显示SQL开关
func (x *XormLoggerAdapter) ShowSQL(show ...bool) {
	if len(show) > 0 {
		x.showSQL = show[0]
	}
}

// IsShowSQL 是否显示SQL
func (x *XormLoggerAdapter) IsShowSQL() bool {
	return x.showSQL
}

// Debug Debug级别日志
func (x *XormLoggerAdapter) Debug(v ...interface{}) {
	x.logger.Debug(v...)
}

// Debugf Debug级别日志
func (x *XormLoggerAdapter) Debugf(format string, v ...interface{}) {
	x.logger.Debugf(format, v...)
}

// Error Error级别日志
func (x *XormLoggerAdapter) Error(v ...interface{}) {
	x.logger.Error(v...)
}

// Errorf Error级别日志
func (x *XormLoggerAdapter) Errorf(format string, v ...interface{}) {
	x.logger.Errorf(format, v...)
}

// Info Info级别日志
func (x *XormLoggerAdapter) Info(v ...interface{}) {
	x.logger.Info(v...)
}

// Infof Info级别日志
func (x *XormLoggerAdapter) Infof(format string, v ...interface{}) {
	// 过滤掉 PING DATABASE 相关的日志
	if strings.Contains(format, "PING DATABASE") {
		return
	}
	x.logger.Infof(format, v...)
}

// Warn Warn级别日志
func (x *XormLoggerAdapter) Warn(v ...interface{}) {
	x.logger.Warn(v...)
}

// Warnf Warn级别日志
func (x *XormLoggerAdapter) Warnf(format string, v ...interface{}) {
	x.logger.Warnf(format, v...)
}

// Debug 便捷函数，直接使用全局Logger
func Debug(args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Debug(args...)
		} else {
			Logger.Debug(args...)
		}
	}
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Debugf(format, args...)
		} else {
			Logger.Debugf(format, args...)
		}
	}
}

// Info ...
func Info(args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Info(args...)
		} else {
			Logger.Info(args...)
		}
	}
}

// Infof ...
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Infof(format, args...)
		} else {
			Logger.Infof(format, args...)
		}
	}
}

// Warn ...
func Warn(args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Warn(args...)
		} else {
			Logger.Warn(args...)
		}
	}
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Warnf(format, args...)
		} else {
			Logger.Warnf(format, args...)
		}
	}
}

// Error ...
func Error(args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Error(args...)
		} else {
			Logger.Error(args...)
		}
	}
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Errorf(format, args...)
		} else {
			Logger.Errorf(format, args...)
		}
	}
}

// Fatal ...
func Fatal(args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Fatal(args...)
		} else {
			Logger.Fatal(args...)
		}
	}
}

// Fatalf ...
func Fatalf(format string, args ...interface{}) {
	if Logger != nil {
		// 获取调用者信息
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Logger.WithField("caller", formatCaller(file, line)).Fatalf(format, args...)
		} else {
			Logger.Fatalf(format, args...)
		}
	}
}

// IgnoreError 静默处理错误，如果有错误则使用Error级别记录到日志但不返回
// 适用于不需要中断程序流程的错误处理场景
func IgnoreError(err error) {
	if err != nil {
		Error("Ignored error:", err)
	}
}

// IgnoreErrorf 静默处理错误并添加自定义消息格式化，使用Error级别记录
// 适用于需要记录特定上下文的错误处理场景
func IgnoreErrorf(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		Errorf("%s: %v", msg, err)
	}
}

// IgnoreErrorWithCallback 静默处理错误、记录日志并执行回调函数
// 适用于需要在错误发生时执行特定逻辑的场景
func IgnoreErrorWithCallback(err error, callback func(error)) {
	if err != nil {
		Error("Ignored error:", err)
		if callback != nil {
			callback(err)
		}
	}
}
