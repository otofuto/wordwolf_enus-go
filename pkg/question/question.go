package question

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"wordwolf_enus/pkg/database"
)

type Question struct {
	Id           int `json:"id"`
	Category     int `json:"category"`
	CategoryName string
	Val1         string `json:"val1"`
	Val2         string `json:"val2"`
}

func All() ([]Question, error) {
	ret := make([]Question, 0)
	db := database.Connect()
	defer db.Close()
	q := "select id, category, val1, val2 from question order by id"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.All db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var qu Question
		err = rows.Scan(&qu.Id, &qu.Category, &qu.Val1, &qu.Val2)
		if err != nil {
			log.Println("question.All rows.Scan")
			return ret, err
		}
		ret = append(ret, qu)
	}
	return ret, nil
}

func List() ([]Question, error) {
	db := database.Connect()
	defer db.Close()
	ret := make([]Question, 0)

	q := "select question.id, category, category.`name`, val1, val2 from question left outer join category on question.category = category.id order by question.id"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.List db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var q Question
		err = rows.Scan(&q.Id, &q.Category, &q.CategoryName, &q.Val1, &q.Val2)
		if err != nil {
			log.Println("question.List rows.Scan")
			return ret, err
		}
		ret = append(ret, q)
	}
	return ret, nil
}

func Search(cate int) ([]Question, error) {
	ret := make([]Question, 0)
	db := database.Connect()
	defer db.Close()
	q := "select id, category, val1, val2 from question where category = " + strconv.Itoa(cate) + " order by id"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.Search db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var qu Question
		err = rows.Scan(&qu.Id, &qu.Category, &qu.Val1, &qu.Val2)
		if err != nil {
			log.Println("question.Search rows.Scan")
			return ret, err
		}
		ret = append(ret, qu)
	}
	return ret, nil
}

func Insert(ques []Question, db *sql.DB) error {
	q := "insert into question (id, category, val1, val2) values "
	for i, que := range ques {
		if i > 0 {
			q += ","
		}
		q += "(" + strconv.Itoa(que.Id) + "," + strconv.Itoa(que.Category) + ",'" + database.Escape(que.Val1) + "','" + database.Escape(que.Val2) + "')"
	}
	ins, err := db.Query(q)
	if err != nil {
		log.Println("question.Insert db.Query")
		return err
	}
	ins.Close()
	return nil
}

func Get(id int) (Question, error) {
	ret := Question{}
	db := database.Connect()
	defer db.Close()

	q := "select id, category, val1, val2 from question where id = " + strconv.Itoa(id)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.Get db.Query")
		return ret, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&ret.Id, &ret.Category, &ret.Val1, &ret.Val2)
		if err != nil {
			log.Println("question.Get rows.Scan")
			return ret, err
		}
	}
	return ret, nil
}

func (que Question) Insert() error {
	db := database.Connect()
	defer db.Close()

	exist, err := checkExist(que.Val1, que.Val2, db)
	if err != nil {
		log.Println("question.Insert checkExist")
		return err
	}
	if exist {
		return errors.New(".既に登録されています")
	}

	q := "select ifnull(max(id), 0) from question"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.Insert db.Query 1")
		return err
	}
	defer rows.Close()
	maxid := 0
	if rows.Next() {
		err = rows.Scan(&maxid)
		if err != nil {
			log.Println("question.Insert rows.Scan")
			return err
		}
	} else {
		return errors.New("????????????")
	}
	maxid++
	q = "insert into question (id, category, val1, val2) values (" + strconv.Itoa(maxid) + "," + strconv.Itoa(que.Category) + ",'" + database.Escape(que.Val1) + "','" + database.Escape(que.Val2) + "')"
	ins, err := db.Query(q)
	if err != nil {
		log.Println("question.Insert db.Query 2")
		return err
	}
	ins.Close()
	return nil
}

func checkExist(val1, val2 string, db *sql.DB) (bool, error) {
	q := "select id from question where val1 = '" + database.Escape(val1) + "' and val2 = '" + database.Escape(val2) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("question.checkExist db.Query")
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (que Question) Update() error {
	db := database.Connect()
	defer db.Close()

	q := "update question set category = " + strconv.Itoa(que.Category) + ", val1 = '" + database.Escape(que.Val1) + "', val2 = '" + database.Escape(que.Val2) + "' where id = " + strconv.Itoa(que.Id)
	upd, err := db.Query(q)
	if err != nil {
		log.Println("question.Update db.Query")
		log.Println(err)
		return err
	}
	upd.Close()
	return nil
}

func Delete(id int) error {
	db := database.Connect()
	defer db.Close()

	q := "delete from question where id = " + strconv.Itoa(id)
	del, err := db.Query(q)
	if err != nil {
		log.Println("question.Delete db.Query")
		return err
	}
	del.Close()
	return nil
}
