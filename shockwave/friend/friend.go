package friend

// Info contains information about a friend.
type Info struct {
	Id         int
	Name       string
	Gender     int
	CustomText string
	Online     bool
	Location   string
	LastAccess string
	FigureData string
}

type Request struct {
	Id   int
	Name string
}
