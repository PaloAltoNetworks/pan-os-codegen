package load

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func Unmarshal(v []byte, o interface{}) error {
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

func File(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return content, err
}
