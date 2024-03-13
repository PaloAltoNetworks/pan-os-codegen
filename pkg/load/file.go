package load

import (
	"os"
)

func File(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return content, err
}
