package account

import (
	"errors"
	"log"
	"net/http"
	"wordwolf_enus/pkg/database"
	"wordwolf_enus/pkg/util"
)

type Account struct {
	Id       int    `json:"id"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func CheckLogin(r *http.Request) Account {
	db := database.Connect()
	defer db.Close()

	a := Account{}
	ck, err := r.Cookie("ww_tk")
	if err != nil {
		return a
	}
	q := "select id, mail from account where token = '" + database.Escape(ck.Value) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("account.CheckLogin db.Query")
		log.Println(err)
		return a
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&a.Id, &a.Mail)
		if err != nil {
			log.Println("account.CheckLogin rows.Scan")
			log.Println(err)
			return a
		}
	}
	return a
}

func Login(mail, pass string) (Account, error) {
	db := database.Connect()
	defer db.Close()

	q := "select id, mail, password from account where mail = '" + database.Escape(mail) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("account.Login db.Query")
		log.Println(err)
		return Account{}, errors.New("アカウントの検索に失敗しました")
	}
	defer rows.Close()
	if rows.Next() {
		var a Account
		err = rows.Scan(&a.Id, &a.Mail, &a.Password)
		if err != nil {
			log.Println("account.Login rows.Scan")
			log.Println(err)
			return Account{}, errors.New("アカウント情報の取得に失敗しました")
		}
		if util.CheckPass(a.Password, pass) {
			a.Token = util.CreateTokenRand(50)
			q = "update account set token = '" + a.Token + "' where mail = '" + database.Escape(mail) + "'"
			upd, err := db.Query(q)
			if err != nil {
				log.Println("account.Login db.Query 2")
				log.Println(err)
				return Account{}, errors.New("ログイン処理に失敗しました")
			}
			upd.Close()
			return a, nil
		}
		return Account{}, errors.New("パスワードが間違っています")
	}
	return Account{}, errors.New("アカウントがみつかりません")
}

func Logout(r *http.Request) {
	db := database.Connect()
	defer db.Close()

	ck, err := r.Cookie("ww_tk")
	if err != nil {
		return
	}
	q := "update account set token = '' where token = '" + ck.Value + "'"
	del, err := db.Query(q)
	if err != nil {
		log.Println("account.Logout db.Query")
		log.Println(err)
		return
	}
	del.Close()
	return
}

func (a Account) Insert() error {
	db := database.Connect()
	defer db.Close()

	q := "select id from account where `mail` = '" + database.Escape(a.Mail) + "'"
	rows, err := db.Query(q)
	if err != nil {
		log.Println("account.Insert db.Query")
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New(".登録済みのメールアドレスです")
	}
	q = "insert into account (`mail`, `password`, `token`) values ('" + database.Escape(a.Mail) + "', '" + util.PassHash(a.Password) + "', '')"
	ins, err := db.Query(q)
	if err != nil {
		log.Println("account.Insert db.Query 2")
		return err
	}
	ins.Close()
	return nil
}
