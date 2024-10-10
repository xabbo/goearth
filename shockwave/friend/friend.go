package friend

// Info contains information about a friend.
type Info struct {
	Id         int
	Name       string
	Gender     int
	Motto string
	Online     bool
	CanFollow bool
	LastAccess string
	Figure string
	CategoryId int
}

type Request struct {
	Id   int
	Name string
}
