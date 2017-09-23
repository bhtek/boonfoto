package main

import (
	"filescanner"
	"log"

	"github.com/rainycape/magick"
	"gopkg.in/mgo.v2"
	"os"
	"bufio"
	"math"
	"fmt"
	"time"
	"gopkg.in/mgo.v2/bson"
)

type MongoPopulator struct {
	coll mgo.Collection
}

func (m MongoPopulator) visitImageFile(path string, modTime time.Time) {
	fmt.Println("imageFile: ", path)
}

type Foto struct {
	Filename string
}


func main() {
	d, err := time.ParseDuration("10s")
	if err != nil {
		log.Fatal("Failed parsing duration: ", err)
	}

	//i := mgo.DialInfo{Database: "boonfotodb", ServiceHost: "ds147044.mlab.com:47044", Username: "pi", Password: "ilovepies", Timeout: d}
	i, err := mgo.ParseURL("pi:ilovepies@ds147044.mlab.com:47044/boonfotodb")
	if err != nil {
		log.Fatal("Failed to parse mongo url: ", err)
	}
	i.Timeout = d

	log.Printf("Initiating connection to Mongo Server.")
	s, err := mgo.DialWithInfo(i)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to connect to mongo: %s", err))
	}
	defer s.Close()

	// Optional. Switch the session to a monotonic behavior.
	s.SetMode(mgo.Monotonic, true)

	c := s.DB("").C("fotos")

	m := MongoPopulator{c}
	filescanner.Scan("/mnt/nas/Pictures/boon-phone-sync/2017", m.visitImageFile)

	//c.Insert(&Foto{"test"})
	//
	//var result []Foto
	//c.Find(bson.M{}).All(&result)
	//for _, v := range result {
	//	fmt.Println("filename: ", v.Filename)
	//}
	//
	//image, err := magick.DecodeFile("t.jpg")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//longer_edge_size := math.Max(float64(image.Width()), float64(image.Height()))
	//multiplier := 200 / longer_edge_size
	//
	//thumbnail_image, err := image.Scale(int(float64(image.Width())*multiplier), int(float64(image.Height())*multiplier))
	//f, err := os.Create("thumb.jpg")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//defer f.Close()
	//
	//w := bufio.NewWriter(f)
	//thumbnail_image.Encode(w, nil)
}
