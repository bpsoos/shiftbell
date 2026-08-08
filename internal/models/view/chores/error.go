package chores

type Link struct {
	Label string
	Href  string
}

type Error struct {
	Message string
	Links   []Link
}
