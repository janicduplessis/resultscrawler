package results

import "github.com/janicduplessis/resultscrawler/pkg/api"

// Store provides an interface for storing results.
type Store interface {
	GetResults(userID string) (*api.Results, error)
	UpdateResults(results *api.Results) error
}
