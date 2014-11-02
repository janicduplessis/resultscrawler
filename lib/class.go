package lib

// Class is an entity for a class.
type Class struct {
	Name    string   `bson:"name"`
	Group   string   `bson:"group"`
	Year    string   `bson:"year"`
	Results []Result `bson:"results"`
}
