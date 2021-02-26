package rosbotcollector

import "context"

type (
	Client interface {
		// ParseWithDefaults returns a slice of Ros-Bot server updates based on the default parsing
		// configuration.
		ParseWithDefaults(ctx context.Context) ([]*ServerUpdate, error)
		// ParseWithConfig returns a slice of Ros-Bot server updates based on the provided
		// parsing configuration.
		ParseWithConfig(ctx context.Context, config *ParserConfig) ([]*ServerUpdate, error)
	}

	client struct {
		httpService HTTPService
	}
)

// NewClient a instance of the `rosbotcollector.Client` interface.
func NewClient(usernameOrEmail string, password string) (Client, error) {
	s, err := newHTTPService(usernameOrEmail, password)
	if err != nil {
		return nil, err
	}
	return &client{httpService: s}, nil
}

func (c *client) ParseWithDefaults(ctx context.Context) ([]*ServerUpdate, error) {
	config := NewParseConfig()
	return newParser(config, c.httpService).Parse(ctx)
}

func (c *client) ParseWithConfig(ctx context.Context, config *ParserConfig) ([]*ServerUpdate, error) {
	return newParser(config, c.httpService).Parse(ctx)
}
