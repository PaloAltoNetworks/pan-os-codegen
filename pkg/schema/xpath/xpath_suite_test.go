package xpathschema_test

import (
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXpath(t *testing.T) {
	handler := slog.NewTextHandler(
		GinkgoWriter,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)
	slog.SetDefault(slog.New(handler))
	RegisterFailHandler(Fail)
	RunSpecs(t, "Xpath Suite")
}
