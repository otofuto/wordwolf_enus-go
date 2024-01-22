package member

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"wordwolf_enus/pkg/database"
)

type Member struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	RoomUrl   string `json:"roomurl"`
	Odai      string `json:"odai"`
	VoteCount int    `json:"votecount"`
}

func (m Member) Insert() (int, error) {
	db := database.Connect()
	defer db.Close()
	q := "insert into member (`name`, roomurl, odai, votecount) values (?, ?, ?, ?)"
	ins, err := db.Prepare(q)
	if err != nil {
		log.Println("member.Insert db.Prepare")
		return 0, err
	}
	defer ins.Close()
	result, err := ins.Exec(m.Name, m.RoomUrl, m.Odai, m.VoteCount)
	if err != nil {
		log.Println("member.Insert ins.Exec")
		return 0, err
	}
	id64, _ := result.LastInsertId()
	m.Id = int(id64)
	return m.Id, nil
}

func Room(hash string, db *sql.DB) ([]Member, error) {
	ret := make([]Member, 0)
	q := "select id, `name`, `roomurl`, odai, votecount from member where roomurl = '" + database.Escape(hash) + "' order by id"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("member.Room db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		err = rows.Scan(&m.Id, &m.Name, &m.RoomUrl, &m.Odai, &m.VoteCount)
		if err != nil {
			log.Println("member.Room rows.Scan")
			return ret, err
		}
		ret = append(ret, m)
	}
	return ret, nil
}

func Get(hash string, id int, db *sql.DB) Member {
	ret := Member{}
	q := "select id, `name`, `roomurl`, odai, votecount from member where roomurl = '" + database.Escape(hash) + "' and id = " + strconv.Itoa(id)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("member.Get db.Query")
		log.Println(err)
		return ret
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&ret.Id, &ret.Name, &ret.RoomUrl, &ret.Odai, &ret.VoteCount)
		if err != nil {
			log.Println("member.Get rows.Scan")
			log.Println(err)
			return ret
		}
	}
	return ret
}

func Voted(hash string, db *sql.DB) (int, error) {
	q := "select ifnull(sum(votecount), 0) from member where roomurl = '" + database.Escape(hash) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("member.Voted db.Query")
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var ret int
		err = rows.Scan(&ret)
		if err != nil {
			log.Println("member.Voted rows.Scan")
			return 0, err
		}
		return ret, nil
	}
	return 0, errors.New("?????????????")
}

func Vote(hash string, oppoid int, db *sql.DB) error {
	q := "update member set votecount = votecount + 1 where id = " + strconv.Itoa(oppoid)
	upd, err := db.Query(q)
	if err != nil {
		log.Println("member.Vote db.Query")
		return err
	}
	upd.Close()
	return nil
}

func CloseRoom(hash string, db *sql.DB) error {
	q := "delete from member where roomurl = '" + database.Escape(hash) + "'"
	del, err := db.Query(q)
	if err != nil {
		log.Println("member.CloseRoom db.Query")
		return err
	}
	del.Close()
	return nil
}

func (m Member) Delete(db *sql.DB) error {
	q := "delete from member where id = " + strconv.Itoa(m.Id)
	del, err := db.Query(q)
	if err != nil {
		log.Println("member.Delete db.Query")
		return err
	}
	del.Close()
	return nil
}

func (m Member) UpdateOdai(db *sql.DB) error {
	q := "update member set odai = '" + database.Escape(m.Odai) + "' where id = " + strconv.Itoa(m.Id)
	upd, err := db.Query(q)
	if err != nil {
		log.Println("member.UpdateOdai db.Query")
		return err
	}
	upd.Close()
	return nil
}

func VoteEnd(hash string) (bool, error) {
	db := database.Connect()
	defer db.Close()

	q := "select (select ifnull(sum(votecount), 0) as sumvote from member where roomurl = '" + hash + "') = playercount from room where `url` = '" + hash + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("member.VoteEnd db.Query")
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var b int
		err = rows.Scan(&b)
		if err != nil {
			log.Println("member.VoteEnd rows.Scan")
			return false, err
		}
		return b == 1, nil
	}
	return false, errors.New("???")
}
