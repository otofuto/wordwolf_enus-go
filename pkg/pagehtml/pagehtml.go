package pagehtml

import (
	"database/sql"
	"html/template"
	"log"
	"wordwolf_enus/pkg/database"
)

type PageHtml struct {
	Id       int           `json:"id"`
	PageName string        `json:"page_name"`
	PagePath string        `json:"page_path"`
	Left     template.HTML `json:"left"`
	Right    template.HTML `json:"right"`
	Top      template.HTML `json:"top"`
	Bottom   template.HTML `json:"bottom"`
}

func Get(pagepath string) PageHtml {
	ret := PageHtml{}
	q := "select id, page_name, page_path, left_html, right_html, top_html, bottom_html from page_html where page_path = '" + database.Escape(pagepath) + "'"
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(q)
	if err != nil {
		log.Println("pagehtml.Get db.Query")
		log.Println(err)
		return ret
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&ret.Id, &ret.PageName, &ret.PagePath, &ret.Left, &ret.Right, &ret.Top, &ret.Bottom)
		if err != nil {
			log.Println("pagehtml.Get rows.Scan")
			log.Println(err)
			return ret
		}
	}
	return ret
}

func All() ([]PageHtml, error) {
	ret := make([]PageHtml, 0)
	q := "select id, page_name, page_path, left_html, right_html, top_html, bottom_html from page_html order by id"
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(q)
	if err != nil {
		log.Println("pagehtml.All db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var ph PageHtml
		err = rows.Scan(&ph.Id, &ph.PageName, &ph.PagePath, &ph.Left, &ph.Right, &ph.Top, &ph.Bottom)
		if err != nil {
			log.Println("pagehtml.All rows.Scan")
			return ret, err
		}
		ret = append(ret, ph)
	}
	return ret, nil
}

func (ph PageHtml) Insert() error {
	db := database.Connect()
	defer db.Close()

	exist, err := checkExist(ph.PagePath, db)
	if err != nil {
		log.Println("pagehtml.Insert checkExist")
		return err
	}
	if exist {
		q := "update pagehtml set page_name = ?, left_html = ?, right_html = ?, top_html = ?, bottom_html = ? where page_path = ?"
		upd, err := db.Prepare(q)
		if err != nil {
			log.Println("pagehtml.Insert db.Prepare: upd")
			return err
		}
		defer upd.Close()
		_, err = upd.Exec(ph.PageName, ph.Left, ph.Right, ph.Top, ph.Bottom, ph.PagePath)
		if err != nil {
			log.Println("pagehtml.Insert upd.Exec")
			return err
		}
		return nil
	}

	q := "insert into pagehtml (page_name, page_path, left_html, right_html, top_html, bottom_html) values (?, ?, ?, ?, ?, ?)"
	ins, err := db.Prepare(q)
	if err != nil {
		log.Println("pagehtml.Insert db.Prepare: ins")
		return err
	}
	defer ins.Close()
	_, err = ins.Exec(ph.PageName, ph.PagePath, ph.Left, ph.Right, ph.Top, ph.Bottom)
	if err != nil {
		log.Println("pagehtml.Insert ins.Exec")
		return err
	}
	return nil
}

func checkExist(path string, db *sql.DB) (bool, error) {
	q := "select id from pagehtml where page_path = '" + database.Escape(path) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("pagehtml.checkExist db.Query")
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}
