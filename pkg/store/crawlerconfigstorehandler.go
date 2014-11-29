package store

import "labix.org/v2/mgo/bson"

const crawlerConfigKey = "crawler_config"

// CrawlerConfigStoreHandler implements the UserInfoStore interface.
type CrawlerConfigStoreHandler struct {
	store Store
}

// NewCrawlerConfigStoreHandler creates a new CrawlerConfigStoreHandler.
func NewCrawlerConfigStoreHandler(store Store) *CrawlerConfigStoreHandler {
	return &CrawlerConfigStoreHandler{
		store: store,
	}
}

// FindByID returns the crawler config for the specified userID.
func (hndl *CrawlerConfigStoreHandler) FindByID(userID bson.ObjectId) (*CrawlerConfig, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	crawlerConfig := CrawlerConfig{}
	err := db.C(crawlerConfigKey).Find(bson.M{"user_id": userID}).One(&crawlerConfig)
	return &crawlerConfig, err
}

// Update updates the crawler config with the specified config.
func (hndl *CrawlerConfigStoreHandler) Update(crawlerConfig *CrawlerConfig) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(crawlerConfigKey).Update(bson.M{"user_id": crawlerConfig.UserID}, crawlerConfig)
	return err
}

// Insert creates a new crawler config.
func (hndl *CrawlerConfigStoreHandler) Insert(crawlerConfig *CrawlerConfig) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(crawlerConfigKey).Insert(crawlerConfig)
	return err
}
