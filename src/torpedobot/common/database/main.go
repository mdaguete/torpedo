package database

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
	Server   string
	Database string
}

type TorpedoStats struct {
	ProcessedMessagesTotal int64
}

func (mdb *MongoDB) GetSession() (session *mgo.Session, err error) {
	session, err = mgo.Dial(mdb.Server)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return
}

func (mdb *MongoDB) GetCollection(collectionName string) (session *mgo.Session, collection *mgo.Collection, err error) {
	session, err = mdb.GetSession()
	if err != nil {
		return
	}
	collection = session.DB(mdb.Database).C(collectionName)
	return
}

func (mdb *MongoDB) GetUpdateTotalMessages(step int64) (count int64){
	session, collection, err := mdb.GetCollection("messagestats")
	if err != nil {
		log.Printf("GetUpdateTotalMessages failed with: %+v\n", err)
		return
	}
	defer session.Close()
	result := TorpedoStats{}
	err = collection.Find(bson.M{}).One(&result)
	if err != nil {
		log.Printf("No stats available: %+v\n", err)
		count = 1
		err = collection.Insert(&result)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		count = result.ProcessedMessagesTotal
	}
	result = TorpedoStats{ProcessedMessagesTotal:count + step}
	err = collection.Update(bson.M{}, result)
	if err != nil {
		log.Printf("Failed to update stats: %+v\n", err)
	}
	return
}

func New(server, database string) (mongodb *MongoDB) {
	if server == "" {
		server = "localhost"
	}
	if database == "" {
		database = "torpedobot"
	}
	mongodb = &MongoDB{Server: server,
		Database: database}
	return
}