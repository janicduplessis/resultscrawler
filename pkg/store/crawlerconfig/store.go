package crawlerconfig

import "github.com/janicduplessis/resultscrawler/pkg/api"

// Store is an interface for storing the crawler configuration.
type Store interface {
	GetCrawlerConfig(userID string) (*api.CrawlerConfig, error)
	UpdateCrawlerConfig(crawlerConfig *api.CrawlerConfig) error
}
