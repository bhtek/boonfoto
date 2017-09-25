package main

import (
	"filescanner"
	"log"

	"gopkg.in/mgo.v2"
	"fmt"
	"time"
	"gopkg.in/mgo.v2/bson"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
)

type MongoPopulator struct {
	coll *mgo.Collection
}

func (m MongoPopulator) visitImageFile(path string, modTime time.Time) {
	c, err := m.coll.Find(bson.M{"path": path}).Count()
	if err != nil {
		log.Fatal("Failed executing find: ", err)
	}

	if c > 0 {
		return
	}

	m.coll.Insert(bson.M{"path": path, "mtime": modTime})
	fmt.Println("Added imageFile: ", path)
}

type SqlPopulator struct {
	db *sql.DB
}

func count(db *sql.DB, query string, args ...interface{}) (int) {
	rows, err := db.Query(query, args ...)
	if err != nil {
		log.Fatal("Failed to get count: ", err)
	}
	defer rows.Close()

	rows.Next()
	var count int
	rows.Scan(&count)
	return count
}

func (sp SqlPopulator) visitImageFile(path string, modTime time.Time) {
	c := count(sp.db, "SELECT COUNT(1) FROM fotos WHERE path = ?", path)
	if c > 0 {
		return
	}

	sp.db.Exec("INSERT INTO fotos (path, mtime) VALUES (?, ?)", path, modTime)
	fmt.Println("Added imageFile: ", path)
}

type Foto struct {
	Filename string
}

func main() {
	fillSqlLiteDb()
}

func fillSqlLiteDb() {
	db, err := sql.Open("sqlite3", "./fotos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)

	sp := SqlPopulator{db}
	filescanner.Scan("/mnt/nas/Pictures/boon-phone-sync/2017", sp.visitImageFile)
}

func createTable(db *sql.DB) {
	rows, err := db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", "fotos")
	if err != nil {
		log.Fatal(err)
	}

	rows.Next()
	var count int
	rows.Scan(&count)
	rows.Close()

	if count < 1 {
		_, err := db.Exec(`
			CREATE TABLE fotos (id INTEGER NOT NULL PRIMARY KEY, path TEXT NOT NULL, mtime DATETIME, rotation INTEGER)
		`)
		if err != nil {
			log.Fatal("Fail to create table: ", err)
		}
		fmt.Println("Created table fotos.")
	}
}

func fillDb() {
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
