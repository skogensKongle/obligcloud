package main

import (
	"fmt"
	"testing"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//var testRates = FromFixer{Base: "EUR", Date: "2017-11-10", Rates: {"BGN": 1.9558000564575195, "USD": 1.1654000282287598, "SEK": 9.743000030517578, "NOK": 9.456000328063965}}
var discord = "https://discordapp.com/api/webhooks/378503095576952842/1tmdzZmVLBHN8DRyGOBHk1N4NTzHR9QaXjzC6eacFYGR7ATsTpeKe4WwQx9S8ZUz6jCK"

func setupDB(t *testing.T) *Mongo {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}

	return &db
}

func tearDownDB(t *testing.T, db *Mongo) {
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}

	err = session.DB(db.DatabaseName).DropDatabase()
	if err != nil {
		t.Error(err)
	}
}

func TestMongo_add(t *testing.T) {
	//db := setupDB(t)
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	db.Init()

	//nr := db.Count()
	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	db.add(data)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.")
	}
}

func TestMongo_get(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	db.Init()
	//nr := db.Count()
	if db.Count() != 0 {
		t.Error("database not properly initialized, data count() should be ", db.Count())
	}

	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	data.ID = bson.NewObjectId()
	db.add(data)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.", db.Count())
	}
	var newData WebHook
	objectid := bson.ObjectId(data.ID).String()
	bson.IsObjectIdHex(objectid)
	fmt.Print(objectid)
	newData = db.get(objectid)

	if newData.Webhookurl != data.Webhookurl ||
		newData.Basecurrency != data.Basecurrency ||
		newData.Targetcurrency != data.Targetcurrency ||
		newData.Mintriggervalue != data.Mintriggervalue ||
		newData.Maxtriggervalue != data.Maxtriggervalue {
		t.Error("data do not match.", newData.ID, newData.Webhookurl, newData.Basecurrency, newData.Targetcurrency, newData.Mintriggervalue, newData.Maxtriggervalue)
	}
}

func TestDel_delete(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	db.Init()
	//nr := db.Count()
	if db.Count() != 0 {
		t.Error("database not properly initialized, data count() should be 0.", db.Count())
	}

	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	data.ID = bson.NewObjectId()
	db.add(data)
	objectid := bson.ObjectId(data.ID).String()
	bson.IsObjectIdHex(objectid)
	fmt.Print(objectid)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.", db.Count())
	}

	db.delete(string(data.ID))
	if db.Count() != 0 {
		t.Error("Could not delete.", db.Count())
	}
}

func TestInter_notInInterval(t *testing.T) {
	var testWeb WebHook
	var testR FromFixer
	testWeb.Targetcurrency = "NOK"
	testWeb.Maxtriggervalue = 2.5
	testWeb.Mintriggervalue = 1.2
	testR.Rates = map[string]float32{"NOK": 1.5}
	if notInInterval(testWeb, testR) {
		t.Error("Cant find interval.")
	}
}
