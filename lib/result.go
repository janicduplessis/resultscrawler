package lib

// Result is an entity for storing a result
type Result struct {
	Name    string `bson:"name"`
	Result  string `bson:"result"`
	Average string `bson:"average"`
}
