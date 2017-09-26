package main

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"log"
	"fmt"
)

func main() {
	db, err := sql.Open("sqlite3", "./fotos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id FROM fotos ORDER BY mtime, path")
	if err != nil {
		log.Fatal("Failed to load fotos: ", err)
	}
	defer rows.Close()

	var idList []int32
	for rows.Next() {
		var id int32
		rows.Scan(&id)
		idList = append(idList, id)
	}
	fmt.Println("Foto id list loaded.")
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
