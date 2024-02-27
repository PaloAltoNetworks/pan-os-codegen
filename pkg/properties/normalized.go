package properties

import (
    "fmt"
    "io/fs"
    "os"
    "runtime"
    "path/filepath"
    "strings"

    "github.com/paloaltonetworks/pan-os-codegen/pkg/load"
)


func GetNormalizations() ([]string, error) {
    _, loc, _, ok := runtime.Caller(0)
    if !ok {
        return nil, fmt.Errorf("couldn't get caller info")
    }

    basePath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(loc))), "specs")

    files := make([]string, 0, 100)

    err := filepath.WalkDir(basePath, func(path string, entry fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if strings.HasSuffix(entry.Name(), ".yaml") {
            files = append(files, path)
        }

        return nil
    })
    if err != nil {
        return nil, err
    }

    return files, nil
}

func ParseSpec(path string) (*Normalization, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var ans Normalization
    err = load.File(b, &ans)
    return &ans, err
}

type Normalization struct {
    Name string `json:"name" yaml:"name"`
    Spec NormalizationSpec `json:"spec" yaml:"spec"`
}

type NormalizationSpec struct {
    Version string `json:"version" yaml:"version"`
}

func (o *Normalization) Sanity() error {
    return nil
}

func (o *Normalization) Validate() []error {
    return nil
}
