package profile

import (
	"strconv"
	"strings"

	g "xabbo.b7c.io/goearth"
)

type Profile struct {
	Name                    string
	Figure                  string
	Gender                  string
	CustomData              string
	PhTickets               int
	PhFigure                string
	PhotoFilm               int
	DirectMail              int
	OnlineStatus            bool
	PublicProfileEnabled    bool
	FriendRequestsEnabled   bool
	OfflineMessagingEnabled bool
}

func (profile *Profile) Parse(p *g.Packet, pos *int) {
	*profile = Profile{}
	for _, line := range strings.Split(p.ReadStringPtr(pos), "\r") {
		kvp := strings.SplitN(line, "=", 2)
		if len(kvp) != 2 {
			dbg.Printf("WARNING: line split length != 2: %q", line)
			continue
		}
		key, val := kvp[0], kvp[1]
		switch key {
		case "name":
			profile.Name = val
		case "figure":
			profile.Figure = val
		case "sex":
			profile.Gender = val
		case "customData":
			profile.CustomData = val
		case "ph_tickets":
			n, err := strconv.Atoi(val)
			if err != nil {
				dbg.Printf("WARNING: invalid integer for ph_tickets: %q", val)
				continue
			}
			profile.PhTickets = n
		case "ph_figure":
			profile.PhFigure = val
		case "photo_film":
			n, err := strconv.Atoi(val)
			if err != nil {
				dbg.Printf("WARNING: invalid integer for photo_film: %q", val)
			}
			profile.PhotoFilm = n
		case "directMail":
			n, err := strconv.Atoi(val)
			if err != nil {
				dbg.Printf("WARNING: invalid integer for direct_mail: %q", val)
			}
			profile.DirectMail = n
		case "onlineStatus":
			profile.OnlineStatus = val == "1"
		case "publicProfileEnabled":
			profile.PublicProfileEnabled = val == "1"
		case "friendRequestsEnabled":
			profile.FriendRequestsEnabled = val == "1"
		case "offlineMessagingEnabled":
			profile.OfflineMessagingEnabled = val == "1"
		}
	}
}
