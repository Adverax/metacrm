package yamlConfig

import (
	"github.com/adverax/metacrm/pkg/access/fetchers/maps/yaml"
	"github.com/adverax/metacrm/pkg/configs"
)

type Source struct {
	fetcher configs.Fetcher
}

func NewFileLoaderBuilder() *configs.FileLoaderBuilder {
	return configs.NewFileLoaderBuilder().
		WithSourceBuilder(
			func(fetcher configs.Fetcher) configs.Source {
				return yamlFetcher.New(fetcher)
			},
		)
}
