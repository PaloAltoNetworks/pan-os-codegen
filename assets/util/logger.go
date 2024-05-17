package util

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// Logger is a struct that contains the configuration for logging.
type Logger struct {
	// Logging is a map of the types of logging to perform.
	// The keys are "send", "receive", and "curl".
	// The values are booleans indicating whether to perform that type of logging.
	// The "send" key indicates whether to log the data being sent.
	// The "receive" key indicates whether to log the data being received.
	// The "curl" key indicates whether to log the data as a curl command.
	// The "personal-data" key indicates whether to log personal data.
	// The "debug" key indicates whether to log debug data.
	Logging map[string]bool

	// Headers is a map of additional headers to include in requests.
	Headers map[string]string

	// SkipVerifyCertificate indicates whether to skip certificate verification when making requests.
	SkipVerifyCertificate bool

	// Hostname is the hostname of the server to connect to.
	Hostname string

	// Protocol is the protocol to use for the connection (e.g., "http", "https").
	Protocol string

	// Port is the port number to use for the connection.
	Port int
}

// LogSend logs the data sent in the specified format.
func (logger *Logger) LogSend(data url.Values) {
	b := &strings.Builder{}

	if logger.Logging["send"] {
		logger.logSendData(b, data)
	}
	if logger.Logging["curl"] {
		logger.logDataAsCurl(b, data)
	}

	if b.Len() > 0 {
		log.Printf("%s", b.String())
	}
}

func (logger *Logger) logSendData(b *strings.Builder, data url.Values) {
	if b.Len() > 0 {
		fmt.Fprintf(b, "\n")
	}
	realSecret := make(map[string]string)
	realSecret["key"] = data.Get("key")
	realSecret["password"] = data.Get("password")
	if realSecret["key"] != "" {
		data.Set("key", "***")
	}
	if realSecret["password"] != "" {
		data.Set("password", "***")
	}
	fmt.Fprintf(b, "Sending data: %#v", data)
	if realSecret["key"] != "" {
		data.Set("key", realSecret["key"])
	}
	if realSecret["password"] != "" {
		data.Set("password", realSecret["password"])
	}
}

func (logger *Logger) logDataAsCurl(b *strings.Builder, data url.Values) {
	if b.Len() > 0 {
		fmt.Fprintf(b, "\n")
	}

	// Separate out the special keys.
	special := map[string]string{
		"key":     "",
		"element": "",
	}
	ev := url.Values{}
	for k := range data {
		var isSpecial bool
		for sk := range special {
			if sk == k {
				isSpecial = true
				special[k] = data.Get(k)
				break
			}
		}
		if !isSpecial {
			ev[k] = make([]string, 0, len(data[k]))
			for i := range data[k] {
				ev[k] = append(ev[k], data[k][i])
			}
		}
	}

	// Build up the curl command.
	fmt.Fprintf(b, "curl")

	// Skip cert verify.
	if logger.SkipVerifyCertificate {
		fmt.Fprintf(b, " -k")
	}
	// Headers.
	if len(logger.Headers) > 0 && logger.Logging["personal-data"] {
		for k, v := range logger.Headers {
			if v != "" {
				fmt.Fprintf(b, " --header '%s: %s'", k, v)
			} else {
				fmt.Fprintf(b, " --header '%s;'", k)
			}
		}
	}
	// Add URL encoded values.
	if special["key"] != "" {
		if logger.Logging["personal-data"] {
			ev.Set("key", special["key"])
		} else {
			ev.Set("key", "APIKEY")
		}
	}
	// Add in the element, if present.
	if special["element"] != "" {
		fmt.Fprintf(b, " --data-urlencode element@element.xml")
	}
	// URL.
	fmt.Fprintf(b, " '%s://", logger.Protocol)
	if logger.Logging["personal-data"] {
		fmt.Fprintf(b, "%s", logger.Hostname)
	} else {
		fmt.Fprintf(b, "HOST")
	}
	if logger.Port != 0 {
		fmt.Fprintf(b, ":%d", logger.Port)
	}
	fmt.Fprintf(b, "/api")
	if len(ev) > 0 {
		fmt.Fprintf(b, "?%s", ev.Encode())
	}
	fmt.Fprintf(b, "'")

	// Data.
	if special["element"] != "" {
		fmt.Fprintf(b, "\nelement.xml:\n%s", special["element"])
	}
}

// LogReceive logs the received data if the "receive" flag is enabled.
func (logger *Logger) LogReceive(body []byte) {
	b := &strings.Builder{}

	if logger.Logging["receive"] {
		logger.logReceiveData(b, body)
	}

	if b.Len() > 0 {
		log.Printf("%s", b.String())
	}
}

func (logger *Logger) logReceiveData(b *strings.Builder, body []byte) {
	if b.Len() > 0 {
		fmt.Fprintf(b, "\n")
	}

	// Replace sensitive content
	replacedBody := string(body)
	startIndex := strings.Index(replacedBody, "<key>")
	endIndex := strings.Index(replacedBody, "</key>")
	if startIndex != -1 && endIndex != -1 {
		replacedBody = replacedBody[:startIndex] + "***" + replacedBody[endIndex+len("</key>"):]
	}

	fmt.Fprintf(b, "Received data: %s", replacedBody)
}

// LogDebug logs the provided data at the debug level if debug logging is enabled.
func (logger *Logger) LogDebug(subject string, data interface{}) {
	if logger.Logging["debug"] {
		log.Printf("Debug [%s]: %v", subject, data)
	}
}
