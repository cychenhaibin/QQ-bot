package newapi

import "context"

type ChannelGroup struct {
	Group          string   `json:"group"`
	Models         []string `json:"models"`
	Availability7d *float64 `json:"availability_7d"`
}

func (c *Client) ListChannels(ctx context.Context) ([]ChannelGroup, error) {
	var response struct {
		Groups []ChannelGroup `json:"groups"`
	}
	if err := c.DoJSON(ctx, "POST", "/api/bot/v1/channels", nil, &response); err != nil {
		return nil, err
	}
	return response.Groups, nil
}
