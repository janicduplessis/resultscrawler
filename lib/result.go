package lib

type Result struct {
	Name    string `bson:"name"`
	Result  string `bson:"result"`
	Average string `bson:"average"`
}
