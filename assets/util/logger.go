package util

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// Replace "path/to" with the actual import path of the XmlApiClient package.

type Logger struct {
	Logging map[string]bool

	Headers               map[string]string
	SkipVerifyCertificate bool
	Hostname              string
	Protocol              string
	Port                  int
}

// logSend logs the data being sent by the XmlApiClient.
// It checks the Logging map to determine which type of logging to perform.
// If Logging["send"] is true, it logs the data using the logSendData method.
// If Logging["curl"] is true, it logs the data as a curl command using the logDataAsCurl method.
// Finally, if there is any logged data, it prints it using log.Printf.
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
	realKey := data.Get("key")
	if realKey != "" {
		data.Set("key", "########")
	}
	fmt.Fprintf(b, "Sending data: %#v", data)
	if realKey != "" {
		data.Set("key", realKey)
	}
}

func (logger *Logger) logDataAsCurl(b *strings.Builder, data url.Values) {
	if b.Len() > 0 {
		fmt.Fprintf(b, "\n")
	}
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

func (logger *Logger) LogReceive(body []byte) {
	b := &strings.Builder{}

	if logger.Logging["receive"] {
		logger.logReceiveData(b, body)
	}

	if b.Len() > 0 {
		log.Printf("%s", b.String())
	}
}

// logReceiveData logs the received data.
func (logger *Logger) logReceiveData(b *strings.Builder, body []byte) {
	if b.Len() > 0 {
		fmt.Fprintf(b, "\n")
	}
	fmt.Fprintf(b, "Received data: %s", body)
}
