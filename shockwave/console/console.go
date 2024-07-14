package console

import "xabbo.b7c.io/goearth/shockwave/friend"

type Init struct {
	PersistentMsg    string
	UserLimit        int
	NormalLimit      int
	ExtendedLimit    int
	Friends          []friend.Info
	Messages         []Message
	CampaignMessages []CampaignMessage
	Requests         []friend.Request
}

type Message struct {
	Id         string
	SenderId   int
	Sender     int
	FigureData string
	Time       string
	Message    string
}

type CampaignMessage struct {
	Id      int
	Url     string
	Link    string
	Message string
}
