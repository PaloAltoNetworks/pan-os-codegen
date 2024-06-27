package pango

import (
	"context"
	"fmt"
	"log/slog"
)

type LogCategory uint

const (
	LogCategoryPango LogCategory = (1 << iota)
	LogCategoryAction
	LogCategoryQuery
	LogCategoryOp
	LogCategoryUid
	LogCategoryLog
	LogCategoryExport
	LogCategoryImport
	LogCategoryXpath
	LogCategorySend
	LogCategoryReceive
	LogCategoryCurl
	LogCategorySensitive
)

var logCategoryToSymbol = map[LogCategory]string{
	LogCategoryPango:     "pango",
	LogCategoryAction:    "action",
	LogCategoryQuery:     "query",
	LogCategoryOp:        "op",
	LogCategoryUid:       "uid",
	LogCategoryLog:       "log",
	LogCategoryExport:    "export",
	LogCategoryImport:    "import",
	LogCategoryXpath:     "xpath",
	LogCategorySend:      "send",
	LogCategoryReceive:   "receive",
	LogCategoryCurl:      "curl",
	LogCategorySensitive: "sensitive",
}

func NewLogCategoryFromSymbols(symbols []string) LogCategory {
	var logCategoriesMask LogCategory
	symbolMap := make(map[string]any)
	for _, sym := range symbols {
		symbolMap[sym] = nil
	}

	for key, value := range logCategoryToSymbol {
		if _, ok := symbolMap[value]; ok {
			logCategoriesMask = logCategoriesMask | key
		}
	}

	return logCategoriesMask
}

func (typ LogCategory) GetLoggingSymbols() []string {
	symbols := make([]string, 0)
	for key, value := range logCategoryToSymbol {
		if typ&key == key {
			symbols = append(symbols, value)
		}
	}
	return symbols
}

// discardHandler is an slog handler which is always disabled and therefore logs nothing.
type discardHandler struct{}

func (d discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (d discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return d }
func (d discardHandler) WithGroup(name string) slog.Handler        { return d }

func LogCategoryToSymbol(typ LogCategory) (string, error) {
	symbol, ok := logCategoryToSymbol[typ]
	if ok {
		return symbol, nil
	}

	return "", fmt.Errorf("Unknown LogCategory: %s", typ)
}

func SymbolToLogCategory(sym string) (LogCategory, error) {
	for key, value := range logCategoryToSymbol {
		if value == sym {
			return key, nil
		}
	}

	return 0, fmt.Errorf("Unknown logging symbol: %s", sym)
}

type selectiveLogger struct {
	logger        *slog.Logger
	discardLogger *slog.Logger
	logMask       LogCategory
}

func newSelectiveLogger(logger *slog.Logger, logMask LogCategory) *selectiveLogger {
	return &selectiveLogger{
		logger:        logger,
		discardLogger: slog.New(discardHandler{}),
		logMask:       logMask,
	}
}

func (l *selectiveLogger) WithLogCategory(typ LogCategory) *slog.Logger {
	category, ok := logCategoryToSymbol[typ]
	if !ok {
		category = "unknown"
	}

	if l.logMask&typ == typ {
		return l.logger.WithGroup(category)
	}
	return l.discardLogger.WithGroup(category)
}

func (l *selectiveLogger) LogMask() LogCategory {
	return l.logMask
}
