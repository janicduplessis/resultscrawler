package lib

type Class struct {
	Name    string   `bson:"name"`
	Group   string   `bson:"group"`
	Year    string   `bson:"year"`
	Results []Result `bson:"results"`
}
