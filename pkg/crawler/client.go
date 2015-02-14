package crawler

import "net/rpc"

// Client implements a rpc client for the crawler.
type Client struct {
	url    string
	client *rpc.Client
}

// NewClient creates a new client
func NewClient(url string) *Client {
	return &Client{url, nil}
}

// Refresh calls the crawler webservice Queue method for the specified
// userID
func (c *Client) Refresh(userID string) error {
	var reply int
	if err := c.doWithRetry("Webservice.Queue", userID, &reply); err != nil {
		return err
	}
	return nil
}

func (c *Client) prepareConnection() error {
	if c.client == nil {
		client, err := rpc.DialHTTP("tcp", c.url)
		if err != nil {
			return err
		}
		c.client = client
	}
	return nil
}

func (c *Client) doWithRetry(method string, args, reply interface{}) error {
	if err := c.prepareConnection(); err != nil {
		return err
	}

	err := c.client.Call(method, args, reply)
	if err == rpc.ErrShutdown {
		c.client = nil
		if err = c.prepareConnection(); err != nil {
			return err
		}
		if err = c.client.Call(method, args, reply); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
