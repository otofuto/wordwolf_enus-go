package blog

import (
	"database/sql"
	"html/template"
	"log"
	"strconv"
	"wordwolf_enus/pkg/database"
)

type Blog struct {
	Id           int           `json:"id"`
	Image        string        `json:"image"`
	Title        string        `json:"title"`
	Content      template.HTML `json:"content"`
	CreatedAt    string        `json:"created_at"`
	ThumbContent string        `json:"thumb_content"`
}

func ThumbContent(b Blog) string {
	ret := ""
	inTag := false
	inQuoart := false
	for _, r := range b.Content {
		if r == '"' {
			inQuoart = !inQuoart
		} else if r == '<' && !inQuoart {
			inTag = true
		} else if r == '>' && !inQuoart {
			inTag = false
		} else if !inTag {
			ret += string(r)
		}
		if len(ret) >= 70 {
			break
		}
	}
	return ret
}

func Page(display_max_count, page int) []Blog {
	ret := make([]Blog, 0)
	db := database.Connect()
	defer db.Close()
	q := "select id, `image`, `title`, `content`, created_at from blog order by id desc limit " + strconv.Itoa(display_max_count) + " offset " + strconv.Itoa(page*display_max_count)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("blog.Page db.Query")
		log.Println(err)
		return ret
	}
	defer rows.Close()
	for rows.Next() {
		var b Blog
		err = rows.Scan(&b.Id, &b.Image, &b.Title, &b.Content, &b.CreatedAt)
		if err != nil {
			log.Println("blog.Page rows.Scan")
			log.Println(err)
			return ret
		}
		b.CreatedAt = b.CreatedAt[:10]
		ret = append(ret, b)
	}
	return ret
}

func Get(id int) Blog {
	db := database.Connect()
	defer db.Close()
	b := Blog{}
	q := "select id, `image`, `title`, `content`, `created_at` from blog where id = " + strconv.Itoa(id)
	rows, err := db.Query(q)
	if err != nil {
		log.Println("blog.Get db.Query")
		log.Println(err)
		return b
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&b.Id, &b.Image, &b.Title, &b.Content, &b.CreatedAt)
		if err != nil {
			log.Println("blog.Get rows.Scan")
			log.Println(err)
			return b
		}
		b.CreatedAt = b.CreatedAt[:10]
	}
	return b
}

func All() ([]Blog, error) {
	ret := make([]Blog, 0)
	db := database.Connect()
	defer db.Close()
	q := "select id, `image`, `title`, `content`, created_at from blog order by id desc"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("blog.Page db.Query")
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var b Blog
		err = rows.Scan(&b.Id, &b.Image, &b.Title, &b.Content, &b.CreatedAt)
		if err != nil {
			log.Println("blog.Page rows.Scan")
			return ret, err
		}
		b.CreatedAt = b.CreatedAt[:10]
		ret = append(ret, b)
	}
	return ret, nil
}

func Delete(id int) error {
	db := database.Connect()
	defer db.Close()

	q := "delete from blog where id = " + strconv.Itoa(id)
	del, err := db.Query(q)
	if err != nil {
		log.Println("blog.Delete db.Query")
		return err
	}
	del.Close()
	return nil
}

func (b Blog) Insert(db *sql.DB) error {
	q := "insert into blog (`image`, `title`, `content`) values (?, ?, ?)"
	ins, err := db.Prepare(q)
	if err != nil {
		log.Println("blog.Insert db.Prepare")
		return err
	}
	defer ins.Close()
	_, err = ins.Exec(b.Image, b.Title, b.Content)
	if err != nil {
		log.Println("blog.Insert ins.Exec")
		return err
	}
	return nil
}

func (b Blog) Update() error {
	db := database.Connect()
	defer db.Close()

	q := "update blog set `image` = ?, `title` = ?, `content` = ? where id = ?"
	upd, err := db.Prepare(q)
	if err != nil {
		log.Println("blog.Update db.Prepare")
		return err
	}
	defer upd.Close()
	_, err = upd.Exec(b.Image, b.Title, b.Content, b.Id)
	if err != nil {
		log.Println("blog.Update ins.Exec")
		return err
	}
	return nil
}
