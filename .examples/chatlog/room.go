package main

import g "xabbo.b7c.io/goearth"

type RoomData struct {
	EnterRoom bool
	RoomInfo
	RoomForward    bool
	StaffPick      bool
	IsGroupMember  bool
	AllInRoomMuted bool
	Moderation     ModerationSettings
	CanMute        bool
	Chat           ChatSettings
}

type RoomInfo struct {
	Id              g.Id
	Name            string
	Owner           g.Id
	OwnerName       string
	Access          int
	Users           int
	MaxUsers        int
	Description     string
	TradePermission int
	Score           int
	Ranking         int
	Category        int
}

type ModerationSettings struct {
	WhoCanMute int
	WhoCanKick int
	WhoCanBan  int
}

type ChatSettings struct {
	Flow                int
	BubbleWidth         int
	ScrollSpeed         int
	TalkHearingDistance int
	FloodProtection     int
}
