package room

import (
	"database/sql"
	"log"
	"strconv"
	"wordwolf_enus/pkg/database"
)

type Room struct {
	Url           string `json:"url"`
	RoomName      string `json:"roomname"`
	PlayerCount   int    `json:"playercount"`
	WordwolfCount int    `json:"wordwolfcount"`
	TalkCategory  int    `json:"talkcategory"`
	TalkTheme     int    `json:"talktheme"`
	OdaiMap       int    `json:"odaimap"`
	TalkTime      int    `json:"talktime"`
	Passcode      string `json:"passcode"`
	CreateDate    string `json:"createdate"`
}

func All() ([]Room, error) {
	db := database.Connect()
	defer db.Close()

	ret := make([]Room, 0)

	q := "select `url`, roomname, playercount, wordwolfcount, talkcategory, talktheme, odaimap, talktime, passcode, createdate from room order by createdate"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("room.All db.Query")
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var r Room
		err = rows.Scan(&r.Url, &r.RoomName, &r.PlayerCount, &r.WordwolfCount, &r.TalkCategory, &r.TalkTheme, &r.OdaiMap, &r.TalkTime, &r.Passcode, &r.CreateDate)
		if err != nil {
			log.Println("room.All rows.Scan")
			return ret, err
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func DeleteOld(h int) error {
	db := database.Connect()
	defer db.Close()

	q := "delete from room where createdate <= subtime(now(), '" + strconv.Itoa(h) + ":00:00')"
	d, err := db.Query(q)
	if err != nil {
		log.Println("room.DeleteOld db.Query")
		return err
	}
	d.Close()
	return nil
}

func Get(hash string, db *sql.DB) (Room, error) {
	ret := Room{}

	q := "select `url`, roomname, playercount, wordwolfcount, talkcategory, talktheme, odaimap, talktime, passcode, createdate from room where `url` = '" + database.Escape(hash) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("room.Get db.Query")
		return ret, err
	}
	defer rows.Close()

	if rows.Next() {
		var r Room
		err = rows.Scan(&r.Url, &r.RoomName, &r.PlayerCount, &r.WordwolfCount, &r.TalkCategory, &r.TalkTheme, &r.OdaiMap, &r.TalkTime, &r.Passcode, &r.CreateDate)
		if err != nil {
			log.Println("room.Get rows.Scan")
			return ret, err
		}
		ret = r
	}
	return ret, nil
}

func (r Room) Insert() error {
	db := database.Connect()
	defer db.Close()
	q := "insert into room (`url`, roomname, playercount, wordwolfcount, talkcategory, talktheme, odaimap, talktime, passcode) values (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	ins, err := db.Prepare(q)
	if err != nil {
		log.Println("room.Insert db.Prepare")
		return err
	}
	defer ins.Close()
	_, err = ins.Exec(r.Url, r.RoomName, r.PlayerCount, r.WordwolfCount, r.TalkCategory, r.TalkTheme, r.OdaiMap, r.TalkTime, r.Passcode)
	if err != nil {
		log.Println("room.Insert ins.Exec")
		return err
	}
	return nil
}

func (r Room) UpdateOdai(db *sql.DB) error {
	q := "update room set talktheme = ?, odaimap = ? where url = ?"
	upd, err := db.Prepare(q)
	if err != nil {
		log.Println("room.UpdateOdai db.Prepare")
		return err
	}
	defer upd.Close()
	_, err = upd.Exec(r.TalkTheme, r.OdaiMap, r.Url)
	if err != nil {
		log.Println("room.UpdateOdai upd.Exec")
		return err
	}
	return nil
}
