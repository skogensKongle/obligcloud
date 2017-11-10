package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/robfig/cron.v2"
)

//Stores details of DB connesction
type Mongo struct {
	DatabaseURL     string
	DatabaseName    string
	MongoCollection string
}

//Getting stuff from fixer.io
type FromFixer struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float32 `json:"rates"`
}

//Storing info given by user
type WebHook struct {
	ID              bson.ObjectId `bson:"_id,omitempty"`
	Webhookurl      string        `json:"webhookURL"`
	Basecurrency    string        `json:"baseCurrency"`
	Targetcurrency  string        `json:"targetCurrency"`
	Mintriggervalue float32       `json:"minTriggerValue"`
	Maxtriggervalue float32       `json:"maxTriggerValue"`
}

type Exsisting struct {
	ID              bson.ObjectId `bson:"_id,omitempty"`
	Basecurrency    string        `json:"baseCurrency"`
	Targetcurrency  string        `json:"targetCurrency"`
	Currentrate     float32       `json:"currentRate"`
	Mintriggervalue float32       `json:"minTriggerValue"`
	Maxtriggervalue float32       `json:"maxTriggerValue"`
}

//var mongoRates = Mongo{DatabaseURL: "127.0.0.1", DatabaseName: "oblig2", MongoCollection: "rates"}
//var mongoWebhooks = Mongo{DatabaseURL: "127.0.0.1", DatabaseName: "oblig2", MongoCollection: "webhooks"}

var mongoRates = Mongo{DatabaseURL: "mongodb://stisoe:1234@ds149855.mlab.com:49855/cloudoblig2", DatabaseName: "cloudoblig2", MongoCollection: "rates"}
var mongoWebhooks = Mongo{DatabaseURL: "mongodb://stisoe:1234@ds149855.mlab.com:49855/cloudoblig2", DatabaseName: "cloudoblig2", MongoCollection: "webhooks"}

//--------------------------------------------------------------------------------------
func main() {

	router := mux.NewRouter()

	//----------------Making the rates update daily------------------------
	cron := cron.New()
	cron.AddFunc("@daily", func() {
		getRates(&mongoRates)
		checkRates(&mongoRates, &mongoWebhooks)
	})
	cron.Start()
	//---------------------------------------------------------------------
	router.HandleFunc("/", handlerpost).Methods("POST")
	router.HandleFunc("/{ID}", handlerEx).Methods("GET")
	router.HandleFunc("/{ID}", handlerDel).Methods("DELETE")
	router.HandleFunc("/average", handlerAver).Methods("POST")

	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), router)
	//err := http.ListenAndServe(":3000", router)
	if err != nil {
		panic(err)
	}
}

//----------------------------------------------------------------------------------------
func (db *Mongo) Init() {
	session, err := mgo.Dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}
	defer session.Close()
}

//---------------------------------------------------------------------------------------
//count DB
func (db *Mongo) Count() int {
	session, err := mgo.Dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}

	//Handle to DB
	count, err := session.DB(db.DatabaseName).C(db.MongoCollection).Count()
	if err != nil {
		fmt.Printf("Error in Count(): %v", err.Error())
		return -1
	}
	return count
}

//----------------------------------------------------------------------------------------
func (db *Mongo) add(new WebHook) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	//Handler to DB
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(new)
	if err != nil {
		fmt.Printf("Error in Insert(): %v", err.Error())
	}
}

//-----------------------------------------------------------------------------------------
func (db *Mongo) Get(keyID string) WebHook {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	webhook := WebHook{}
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Find(bson.M{"_id": bson.ObjectIdHex(keyID)}).One(&webhook)
	if err != nil {
		return webhook
	}
	return webhook
}

//---------------------------------------------------------------------------------------
func (db *Mongo) Delete(keyID string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.DB(db.DatabaseName).C(db.MongoCollection).RemoveId(bson.ObjectIdHex(keyID))
}

//---------------------------------------------------------------------------------------

func checkRates(db *Mongo, dc *Mongo) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	var rates FromFixer
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Find(nil).Sort("-_id").One(&rates)
	if err != nil {
		fmt.Print("Things are not right")
	}

	var webhooks []WebHook
	session.DB(dc.DatabaseName).C(dc.MongoCollection).Find(nil).All(&webhooks)

	for _, webhook := range webhooks {
		if notInInterval(webhook, rates) {
			var exs Exsisting
			exs.ID = webhook.ID
			exs.Basecurrency = rates.Base
			exs.Targetcurrency = webhook.Targetcurrency
			exs.Currentrate = float32(1 / rates.Rates[webhook.Targetcurrency])
			exs.Mintriggervalue = webhook.Mintriggervalue
			exs.Maxtriggervalue = webhook.Maxtriggervalue
			var url string = "https://discordapp.com/api/webhooks/378503095576952842/1tmdzZmVLBHN8DRyGOBHk1N4NTzHR9QaXjzC6eacFYGR7ATsTpeKe4WwQx9S8ZUz6jCK"
			jsonValue, _ := json.Marshal(exs)
			http.Post(url, "application/json", bytes.NewBuffer(jsonValue))

		}
	}

}

//--------------------------------------------------------------------------------------

func notInInterval(wb WebHook, r FromFixer) bool {
	return !(wb.Mintriggervalue < r.Rates[wb.Targetcurrency] && wb.Maxtriggervalue > r.Rates[wb.Targetcurrency])
}

//--------------------------------------------------------------------------------------

func getRates(db *Mongo) {

	var getAllRates FromFixer

	fmt.Print(db.DatabaseURL)
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	//Fetch response form url
	var url = "https://api.fixer.io/latest"

	repo, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer repo.Body.Close()

	//Grab body from Response

	body, err := ioutil.ReadAll(repo.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &getAllRates)
	if err != nil {
		panic(err)
	}

	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(getAllRates)
	if err != nil {
		fmt.Printf("Error in Insert(): %v", err.Error())
	}
}

//----------------------------------------------------------------------------------------

//----------------------------------------------------------------------------------------

func handlerpost(res http.ResponseWriter, req *http.Request) {

	var webHook WebHook
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&webHook)
	if err != nil {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	webHook.ID = bson.NewObjectId()
	mongoWebhooks.add(webHook)
	//Returne response code
	res.WriteHeader(http.StatusOK)
	fmt.Fprintln(res, webHook.ID.Hex())
}

func handlerEx(res http.ResponseWriter, req *http.Request) {

	ting := mux.Vars(req)
	webshit := mongoWebhooks.Get(ting["ID"])
	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res, webshit)
}

func handlerDel(res http.ResponseWriter, req *http.Request) {
	ting := mux.Vars(req)
	mongoWebhooks.Delete(ting["ID"])
	res.WriteHeader(http.StatusOK)
}

func handlerAver(res http.ResponseWriter, req *http.Request) {
	var webhook WebHook
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&webhook)
	if err != nil {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	session, err := mgo.Dial(mongoRates.DatabaseURL)
	if err != nil {
		panic(err)
	}
	var rates []FromFixer
	err = session.DB(mongoRates.DatabaseName).C(mongoRates.MongoCollection).Find(nil).Sort("-_id").Limit(7).All(&rates)
	if err != nil {
		fmt.Print("Things are not right")
	}
	var days float32 = 0
	for _, rate := range rates {
		days += rate.Rates[webhook.Targetcurrency]
	}
	fmt.Fprint(res, days/float32(len(rates)))

}
