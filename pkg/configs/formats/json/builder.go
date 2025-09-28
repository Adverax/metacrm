package jsonConfig

import (
	"github.com/adverax/metacrm/pkg/access/fetchers/maps/json"
	"github.com/adverax/metacrm/pkg/configs"
)

func NewFileLoaderBuilder() *configs.FileLoaderBuilder {
	return configs.NewFileLoaderBuilder().
		WithSourceBuilder(
			func(fetcher configs.Fetcher) configs.Source {
				return jsonFetcher.New(fetcher)
			},
		)
}
