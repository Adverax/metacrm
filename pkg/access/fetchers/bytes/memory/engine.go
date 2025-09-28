package memoryFetcher

type Fetcher struct {
	data []byte
}

func New(data []byte) *Fetcher {
	return &Fetcher{
		data: data,
	}
}

func (that *Fetcher) Fetch() ([]byte, error) {
	return that.data, nil
}
