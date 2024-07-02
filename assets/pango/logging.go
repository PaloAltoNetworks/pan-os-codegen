package pango

import (
	"context"
	"fmt"
	"log/slog"
	"math/bits"
)

// LogCategory is a bitmask describing what categories of logging to enable.
//
// Available bit-wise flags are as follows:
//
//   - LogCategoryPango: Basic library-level logging
//   - LogCategoryOp: Logging of operation commands (Op functions)
//   - LogCategorySend: Logging of the data being sent to the server. All sensitive
//     data will be scrubbed from the response unless LogCategorySensitive
//     is explicitly added to the mask
//   - LogCategoryReceive: Logging of the data being received from the server. All
//     sensitive data will be scrubbed from the response unless LogCategorySensitive
//     is explicitly added to the mask
//   - LogCategoryCurl: When used along with LogCategorySend, an equivalent curl
//     command will be logged
//   - LogCategoryAll: A meta-category, enabling all above categories at once
//   - LogCategorySensitive: Logging of sensitive data like hostnames, logins,
//     passwords or API keys of any sort
type LogCategory uint

const (
	LogCategoryPango LogCategory = (1 << iota)
	LogCategoryOp
	LogCategorySend
	LogCategoryReceive
	LogCategoryCurl
	LogCategoryAll = LogCategoryPango | LogCategoryOp | LogCategorySend |
		LogCategoryReceive | LogCategoryCurl
	// Make sure that LogCategorySensitive is always last, explicitly set to 1 << 32
	LogCategorySensitive LogCategory = 1 << 32
)

var logCategoryToString = map[LogCategory]string{
	LogCategoryOp:        "op",
	LogCategoryPango:     "pango",
	LogCategorySend:      "send",
	LogCategoryReceive:   "receive",
	LogCategoryCurl:      "curl",
	LogCategoryAll:       "all",
	LogCategorySensitive: "sensitive",
}

// LogCategoryFromStrings transforms list with categories into its bitmask equivalent.
//
// This function takes a list of strings, representing log categories
// that can be used to filter what gets logged by pango library. This list
// can change over time as more categories are added to the library.
//
// It returns LogCategory bitmask which can be then used to configure
// logger. If unknown log category string is given as part of the
// list, error is returned instead.
func LogCategoryFromStrings(symbols []string) (LogCategory, error) {
	// Instead of keeping two maps for two way association, we
	// just generate reversed map on the fly. This function is not
	// going to be used outside of the initial library setup, so
	// the slight performance penalty is not an issue.
	symbolMap := make(map[string]any)
	for _, sym := range symbols {
		symbolMap[sym] = nil
	}

	// Iterate over all keys in logCategoryToString to match
	// category strings to their bitmask representation. Even if
	// LogCategoryAll is matched, we still continue to check for
	// the LogCategorySensitive presence.
	var logCategoriesMask LogCategory
	for key, value := range logCategoryToString {
		if _, ok := symbolMap[value]; ok {
			logCategoriesMask = logCategoriesMask | key
		} else {
			return 0, fmt.Errorf("unknown log category: %s", value)
		}
	}

	return logCategoriesMask
}

// LogCategoryAsStrings interprets given LogCategory bitmask into its string representation.
//
// This function takes LogCategory bitmask as argument, and converts
// it into a list of strings, where each element represents a single
// category. LogCategoryAll is converted into a list of enabled
// categories, without "all".
//
// It returns a list of categories as strings, or error if invalid
// LogCategory mask has been provided.
func (categories LogCategory) LogCategoryAsStrings() ([]string, error) {
	symbols := make([]string, 0)

	// Calculate a number of high bits in the categories mask, to make
	// sure all categories other than LogCategoryAll have been matched.
	highBits := bits.OnesCount(categories)

	// Iterate over all available log categories, skipping
	// LogCategoryAll as we can't distinguish between explicitly
	// ORing all LogCategories and using LogCategoryAll.
	for key, value := range logCategoryToString {
		if key == LogCategoryAll {
			continue
		}
		if categories&key == key {
			symbols = append(symbols, value)
		}
	}

	// Return an error if number of high bits in the categories
	// mask is lower than length of the symbols list
	if len(symbols) < highBits && (categories&LogCategoryAll == LogCategoryAll) {
		return nil, fmt.Errorf("invalid LogCategory mask")
	}

	return symbols
}

// LogCategoryToSymbol returns string representation of the given LogCategory
//
// The given LogCategory can only have single bit set high, and cannot
// match LogCategoryAll. To convert LogCategory bitmask into a list of categories,
// use LogCategoryToStrings instead.
//
// It returns string representation of the log category, or error if
// unknown category has been provided.
func LogCategoryToString(category LogCategory) (string, error) {
	if category&LogCategoryAll == LogCategoryAll {
		return "", fmt.Errorf("cannot convert LogCategoryAll into a category string.")
	}
	symbol, ok := logCategoryToString[typ]
	if ok {
		return symbol, nil
	}

	return "", fmt.Errorf("unknown LogCategory: %s", typ)
}

// StringToLogCategory returns LogCategory mask matching given string category.
//
// The given string should be a single category, and not "all". To convert "all"
// into a list of enabled log categories, use LogCategoryFromStrings.
//
// It returns LogCategory representation of the given category string, or en
// error if either "all" or unknown string has been given.
func StringToLogCategory(sym string) (LogCategory, error) {
	if sym == logCategoryToString[LogCategoryAll] {
		return 0, fmt.Errorf("cannot convert \"all\" category string into LogCategory")
	}
	for key, value := range logCategoryToString {
		if value == sym {
			return key, nil
		}
	}

	return 0, fmt.Errorf("Unknown logging symbol: %s", sym)
}

type categoryLogger struct {
	logger        *slog.Logger
	discardLogger *slog.Logger
	categories    LogCategory
}

func newCategoryLogger(logger *slog.Logger, categories LogCategory) *categoryLogger {
	return &categoryLogger{
		logger:        logger,
		discardLogger: slog.New(discardHandler{}),
		categories:    categories,
	}
}

func (l *categoryLogger) WithLogCategory(category LogCategory) *slog.Logger {
	category, ok := logCategoryToString[category]
	if !ok {
		category = "unknown"
	}

	if l.categories&category == category {
		return l.logger.WithGroup(category)
	}
	return l.discardLogger.WithGroup(category)
}

func (l *categoryLogger) logCategories() LogCategory {
	return l.categories
}
