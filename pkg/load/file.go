package load

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func File(v []byte, o interface{}) error {
	var err error

	runes := []byte(strings.TrimSpace(string(v)))
	if len(runes) == 0 {
		return fmt.Errorf("no data in file")
	}

	if runes[0] == '{' && runes[len(runes)-1] == '}' {
		err = json.Unmarshal(v, o)
	} else {
		err = yaml.Unmarshal(v, o)
	}

	return err
}

func FileContent(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return content, err
}
