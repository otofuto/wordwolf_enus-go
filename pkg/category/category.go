package category

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"wordwolf_enus/pkg/database"
)

type Category struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func All() ([]Category, error) {
	db := database.Connect()
	defer db.Close()

	ret := make([]Category, 0)

	q := "select id, `name` from category order by id"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("category.All db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var c Category
		err = rows.Scan(&c.Id, &c.Name)
		if err != nil {
			log.Println("category.All rows.Scan")
			return ret, err
		}
		ret = append(ret, c)
	}
	return ret, nil
}

func Insert(cates []Category, db *sql.DB) error {
	q := "insert into category (id, `name`) values "
	for i, c := range cates {
		if i > 0 {
			q += ","
		}
		q += "(" + strconv.Itoa(c.Id) + ",'" + database.Escape(c.Name) + "')"
	}
	ins, err := db.Query(q)
	if err != nil {
		log.Println("category.Insert db.Query")
		return err
	}
	ins.Close()
	return nil
}

func Get(id int) (Category, error) {
	db := database.Connect()
	defer db.Close()

	q := "select id, `name` from category where id = " + strconv.Itoa(id)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("category.Get db.Query")
		return Category{}, err
	}
	defer rows.Close()
	var cate Category
	if rows.Next() {
		err = rows.Scan(&cate.Id, &cate.Name)
		if err != nil {
			log.Println("category.Get rows.Scan")
			return Category{}, err
		}
	}
	return cate, nil
}

func (c Category) Insert() error {
	db := database.Connect()
	defer db.Close()

	exist, err := checkExist(c.Name, db)
	if err != nil {
		log.Println("category.Insert checkExist")
		return err
	}
	if exist {
		return errors.New(".登録済みです")
	}

	q := "select ifnull(max(id), 0) from category"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("category.Insert db.Query")
		return err
	}
	defer rows.Close()
	maxid := 0
	if rows.Next() {
		err = rows.Scan(&maxid)
		if err != nil {
			log.Println("category.Insert rows.Scan")
			return err
		}
	}
	maxid++
	q = "insert into category (id, `name`) values (" + strconv.Itoa(maxid) + ",'" + database.Escape(c.Name) + "')"
	ins, err := db.Query(q)
	if err != nil {
		log.Println("category.Insert db.Query")
		return err
	}
	ins.Close()
	return nil
}

func checkExist(name string, db *sql.DB) (bool, error) {
	q := "select id from category where `name` = '" + database.Escape(name) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("category.checkExist db.Query")
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (c Category) Update() error {
	db := database.Connect()
	defer db.Close()

	exist, err := checkExist2(c.Id, db)
	if err != nil {
		log.Println("category.Update checkExist2")
		return err
	}
	if !exist {
		return errors.New("存在しないIDです")
	}

	q := "update category set `name` = '" + database.Escape(c.Name) + "' where id = " + strconv.Itoa(c.Id)
	upd, err := db.Query(q)
	if err != nil {
		log.Println("categiry.Update db.Query")
		return err
	}
	upd.Close()
	return nil
}

func checkExist2(id int, db *sql.DB) (bool, error) {
	q := "select id from category where id = " + strconv.Itoa(id)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("category.checkExist2 db.Query")
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func Delete(id int) error {
	db := database.Connect()
	defer db.Close()

	q := "delete from category where id = " + strconv.Itoa(id)
	del, err := db.Query(q)
	if err != nil {
		log.Println("category.Delete db.Query")
		return err
	}
	del.Close()
	return nil
}
