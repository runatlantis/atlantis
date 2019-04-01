package fetchers

type FetcherConfig struct {
	ConfigType   ConfigSourceType
	GithubConfig *GithubFetcherConfig
}

type Fetcher interface {
	FetchConfig() (string, error)
}
