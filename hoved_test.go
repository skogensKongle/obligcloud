package main

import "testing"


//Stores details of DB connesction
type TestMongo struct {
	DatabaseURL     string
	DatabaseName    string
	MongoCollection string
}

//Getting stuff from fixer.io
type TestFromFixer struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float32 `json:"rates"`
}

//Storing info given by user
type TestWebHook struct {
	ID              bson.ObjectId `bson:"_id,omitempty"`
	Webhookurl      string        `json:"webhookURL"`
	Basecurrency    string        `json:"baseCurrency"`
	Targetcurrency  string        `json:"targetCurrency"`
	Mintriggervalue float32       `json:"minTriggerValue"`
	Maxtriggervalue float32       `json:"maxTriggerValue"`
}

type TestExsisting struct {
	ID              bson.ObjectId `bson:"_id,omitempty"`
	Basecurrency    string        `json:"baseCurrency"`
	Targetcurrency  string        `json:"targetCurrency"`
	Currentrate     float32       `json:"currentRate"`
	Mintriggervalue float32       `json:"minTriggerValue"`
	Maxtriggervalue float32       `json:"maxTriggerValue"`
}
