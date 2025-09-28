package fileFetcher

import (
	"io"
	"os"
)

type Fetcher struct {
	filename   string
	mustExists bool
}

func (that *Fetcher) Fetch() ([]byte, error) {
	file, err := os.Open(that.filename)
	if err != nil {
		if os.IsNotExist(err) && !that.mustExists {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func (that *Fetcher) Save(data []byte) error {
	return os.WriteFile(that.filename, data, 0644)
}
