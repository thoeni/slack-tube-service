package main

type slackRequest struct {
	Token       string   `schema:"token"`
	TeamID      string   `schema:"team_id"`
	TeamDomain  string   `schema:"team_domain"`
	ChannelID   string   `schema:"channel_id"`
	ChannelName string   `schema:"channel_name"`
	UserID      string   `schema:"user_id"`
	Username    string   `schema:"user_name"`
	Command     string   `schema:"command"`
	Text        []string `schema:"text"`
	ResponseURL string   `schema:"response_url"`
}
