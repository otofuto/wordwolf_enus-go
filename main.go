package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"wordwolf_enus/pkg/account"
	"wordwolf_enus/pkg/blog"
	"wordwolf_enus/pkg/category"
	"wordwolf_enus/pkg/database"
	"wordwolf_enus/pkg/member"
	"wordwolf_enus/pkg/pagehtml"
	"wordwolf_enus/pkg/question"
	"wordwolf_enus/pkg/room"
	"wordwolf_enus/pkg/util"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme/autocert"
)

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan SocketMessage)
var upgrader = websocket.Upgrader{}

type SocketMessage struct {
	RoomId  string
	Id      int    `json:"id"`
	Url     string `json:"url"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

type TempContext struct {
	Page         int
	PageLength   int
	Hash         string
	Host         string
	Query        string
	Sort         string
	ReturnPath   string
	UserAgent    string
	Json         string
	Category     category.Category
	CategoryList []category.Category
	Question     question.Question
	QuestionList []question.Question
	Message      string
	RoomName     string
	RoomnameOg   string
	PH           pagehtml.PageHtml
	Pages        []pagehtml.PageHtml
	Room         room.Room
	MyId         int
	UserName     string
	Odai         string
	Sec          int
	Members      []member.Member
	PlayerCount  int
	Blog         blog.Blog
	BlogList     []blog.Blog
	Login        account.Account
	Mode         string
	Len          interface{}
}

type ResultList struct {
	ResultType int         `json:"result_type"`
	List       interface{} `json:"list"`
	Members    interface{} `json:"members"`
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if len(os.Args) > 1 {
		port = os.Args[1]
		if port == "ssl" {
			port = "443"
		}
	}
	if port == "" {
		port = "5000"
	}

	mux := http.NewServeMux()
	mux.Handle("/st/", http.StripPrefix("/st/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/", IndexHandle)
	mux.HandleFunc("/share", ShareHandle)
	mux.HandleFunc("/blog/", BlogHandle)
	mux.HandleFunc("/manage/", ManageHandle)
	mux.HandleFunc("/r/", ApiHandle)
	mux.HandleFunc("/nohup.out", OutHandle)
	mux.HandleFunc("/sw.js", SwjsHandle)
	mux.HandleFunc("/favicon.ico", FaviconHandle)
	mux.HandleFunc("/robots.txt", RobotsHandle)
	mux.HandleFunc("/sitemap.xml", SiteMapHandle)
	mux.HandleFunc("/ws/", SocketHandle)
	go handleMessages()
	log.Println("Listening on port: " + port)
	if port == "443" {
		log.Println("SSL")
		go func() {
			mux2 := http.NewServeMux()
			mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+os.Getenv("DOMAIN")+r.URL.Path, 301)
			})
			if err := http.ListenAndServe(":80", mux2); err != nil {
				panic(err)
			}
		}()
		if err := http.Serve(autocert.NewListener(os.Getenv("DOMAIN")), mux); err != nil {
			panic(err)
		}
	} else if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}

func IndexHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodGet {
		context := TempContext{
			UserAgent: r.UserAgent(),
			Host:      os.Getenv("HOST"),
			Query:     "20240117",
		}
		mode := ""
		hash := ""
		if len(r.URL.Path) > 1 {
			mode = r.URL.Path[1:]
			if strings.Index(mode, "/") > 0 {
				hash = mode[strings.LastIndex(mode, "/")+1:]
				mode = mode[:strings.LastIndex(mode, "/")]
			}
		}
		context.Hash = hash
		filename := ""
		if !util.CheckRequest(w, r) {
			return
		}
		var err error
		if r.URL.Path == "/" {
			filename = "index"
			context.PH = pagehtml.Get("index")
		} else if mode == "setting" {
			err := room.DeleteOld(24 * 3)
			if err != nil {
				log.Println("IndexHandle room.DeleteOld")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			filename = "setting"
			list, err := category.All()
			if err != nil {
				log.Println("IndexHandle category.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			context.CategoryList = list
		} else if mode == "waiting" {
			filename = "waiting"
			db := database.Connect()
			defer db.Close()
			context.Room, err = room.Get(hash, db)
			if err != nil {
				log.Println("IndexHandle waiting room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if context.Room.Url == "" {
				context.Hash = ""
			} else {
				myid, err := strconv.Atoi(r.FormValue("id"))
				if err != nil {
					Page404(w)
					return
				}
				mem := member.Get(hash, myid, db)
				if mem.Id == 0 {
					context.Hash = ""
				}
			}
		} else if mode == "play" {
			filename = "play"
			http.SetCookie(w, &http.Cookie{
				Name:     "ww_voted",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600 * 24 * 3,
				Value:    "false",
			})
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("IndexHandle play room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if rm.Url != "" {
				ck, err := r.Cookie("ww_myname")
				if err != nil || ck.Value == "" {
					context.Odai = "Lost Session"
					context.Hash = ""
				} else {
					context.Sec = rm.TalkTime
					context.Odai = member.Get(hash, toInt(r.FormValue("id")), db).Odai
				}
			} else {
				context.Odai = "This room has been deleted."
				context.Hash = ""
			}
		} else if mode == "vote" {
			filename = "vote"
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("IndexHandle vote room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if rm.Url != "" {
				context.Members, err = member.Room(hash, db)
				if err != nil {
					log.Println("IndexHandle vote member.Room")
					log.Println(err)
					Page500(w, err.Error())
					return
				}
			} else {
				context.Hash = ""
			}
		} else if mode == "counting" {
			filename = "counting"
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("IndexHandle counting room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if rm.Url != "" {
				context.PlayerCount = rm.PlayerCount

			} else {
				context.Hash = ""
				context.PlayerCount = 0
			}
		} else if mode == "announce" {
			filename = "announce"
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("IndexHandle accounce room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if rm.Url == "" {
				Page404(w)
				return
			}
		} else if mode == "privacypolicy" {
			filename = "text"
		} else if mode == "howto" || mode == "about" || mode == "odai" || mode == "questions" {
			filename = mode
			context.PH = pagehtml.Get(mode)
		} else if mode == "offsetting" {
			filename = "offsetting"
			context.CategoryList, err = category.All()
			if err != nil {
				log.Println("IndexHandle offsetting category.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else if mode == "name" || mode == "game" || mode == "offplay" || mode == "finish" || mode == "offannounce" || mode == "categorysetting" {
			filename = mode
		} else {
			filename = "room"
			context.Hash = mode
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(mode, db)
			if err != nil {
				log.Println("IndexHandle room.Get")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			if rm.Url == "" {
				context.RoomnameOg = "This room has been closed."
			} else {
				context.RoomnameOg = rm.RoomName
				context.RoomName = rm.RoomName
			}
		}
		if err := template.Must(template.ParseFiles("template/app/"+filename+".html")).Execute(w, context); err != nil {
			log.Println(err)
			http.Error(w, "500", 500)
			return
		}
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func ShareHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodPost {
		context := TempContext{
			UserAgent: r.UserAgent(),
			Host:      os.Getenv("HOST"),
			Query:     "20240117",
		}
		if !util.CheckRequest(w, r) {
			return
		}
		r.ParseMultipartForm(32 << 20)
		exist := 0
		for _, key := range []string{"roomname", "playercount", "wordwolfcount", "talkcategory", "talktime", "passcode", "username"} {
			if r.FormValue(key) != "" {
				exist++
			}
		}
		if exist == 7 {
			// 議論時間を時:分:秒から秒に変換します。
			talktime := 0
			talktime_list := strings.Split(r.FormValue("talktime"), ":")
			talktime += toInt(talktime_list[0]) * 3600
			talktime += toInt(talktime_list[1]) * 60
			if len(talktime_list) == 3 {
				talktime += toInt(talktime_list[2])
			}
			// 選択されたテーマ(=カテゴリ)からお題をランダムで決めます
			q, err := question.Search(toInt(r.FormValue("talkcategory")))
			if err != nil {
				log.Println("ShareHandle question.Search")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			rand.Seed(time.Now().UnixNano())
			rnd := rand.Intn(len(q) - 1)
			// メンバーに割り振るお題を決めます。2進数で0ならval1、1ならval2をログイン順に割り振り
			omap := append(makeStr("0", toInt(r.FormValue("playercount"))-toInt(r.FormValue("wordwolfcount"))), makeStr("1", toInt(r.FormValue("wordwolfcount")))...)
			rand.Seed(time.Now().UnixNano() - 123456)
			rand.Shuffle(len(omap), func(i, j int) {
				omap[i], omap[j] = omap[j], omap[i]
			})
			omap_int, err := strconv.ParseInt(strings.Join(omap, ""), 2, 64)
			if err != nil {
				log.Println("ShareHandle strconv.ParseInt")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			rm := room.Room{
				Url:           util.GetSha256String(strconv.Itoa(int(time.Now().UnixNano())) + r.FormValue("roomname"))[:10],
				RoomName:      r.FormValue("roomname"),
				PlayerCount:   toInt(r.FormValue("playercount")),
				WordwolfCount: toInt(r.FormValue("wordwolfcount")),
				TalkCategory:  toInt(r.FormValue("talkcategory")),
				TalkTheme:     q[rnd].Id,
				OdaiMap:       int(omap_int),
				TalkTime:      talktime,
				Passcode:      r.FormValue("passcode"),
			}
			err = rm.Insert()
			if err != nil {
				log.Println("ShareHandle room.Insert")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			// 部屋作成者のお題を決定
			odai := ""
			if omap[0] == "0" {
				odai = q[rnd].Val1
			} else {
				odai = q[rnd].Val2
			}
			mem := member.Member{
				Name:    r.FormValue("username"),
				RoomUrl: rm.Url,
				Odai:    odai,
			}
			newid, err := mem.Insert()
			if err != nil {
				log.Println("ShareHandle mem.Insert()")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			mem.Id = newid
			http.SetCookie(w, &http.Cookie{
				Name:     "ww_myname",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600 * 24 * 3,
				Value:    mem.Name,
			})
			http.SetCookie(w, &http.Cookie{
				Name:     "ww_myid",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600 * 24 * 3,
				Value:    strconv.Itoa(mem.Id),
			})
			context.Hash = rm.Url
			context.MyId = mem.Id
			context.UserName = mem.Name
		} else {
			Page404(w)
			return
		}
		if err := template.Must(template.ParseFiles("template/app/share.html")).Execute(w, context); err != nil {
			log.Println(err)
			http.Error(w, "500", 500)
			return
		}
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func toInt(str string) int {
	ret, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return ret
}

func makeStr(chr string, cnt int) []string {
	ret := make([]string, 0)
	for i := 0; i < cnt; i++ {
		ret = append(ret, chr)
	}
	return ret
}

func BlogHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodGet {
		context := TempContext{
			UserAgent: r.UserAgent(),
			Host:      os.Getenv("HOST"),
			Query:     "20240117",
			Len: func(list []blog.Blog) int {
				return len(list)
			},
		}
		mode := ""
		hash := ""
		if len(r.URL.Path) > len("/blog/") {
			mode = r.URL.Path[len("/blog/"):]
			if strings.Index(mode, "/") > 0 {
				hash = mode[strings.LastIndex(mode, "/")+1:]
				mode = mode[:strings.LastIndex(mode, "/")]
			}
		}
		context.Hash = hash
		filename := ""
		if !util.CheckRequest(w, r) {
			return
		}
		if mode == "" {
			filename = "index"
			context.PH = pagehtml.Get("blogindex")
			display_max_count := 20
			context.BlogList = blog.Page(display_max_count, 0)
			for i := 0; i < len(context.BlogList); i++ {
				context.BlogList[i].ThumbContent = blog.ThumbContent(context.BlogList[i])
			}
		} else {
			filename = "blog"
			context.Blog = blog.Get(toInt(mode))
		}
		if err := template.Must(template.ParseFiles("template/blog/"+filename+".html")).Execute(w, context); err != nil {
			log.Println(err)
			http.Error(w, "500", 500)
			return
		}
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func ManageHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodGet {
		context := TempContext{
			UserAgent: r.UserAgent(),
			Host:      os.Getenv("HOST"),
			Query:     "20240117",
			Login:     account.CheckLogin(r),
			Len: func(list []interface{}) int {
				return len(list)
			},
		}
		mode := ""
		hash := ""
		if len(r.URL.Path) > len("/manage/") {
			mode = r.URL.Path[len("/manage/"):]
			if strings.Index(mode, "/") > 0 {
				hash = mode[strings.LastIndex(mode, "/")+1:]
				mode = mode[:strings.LastIndex(mode, "/")]
			}
		}
		context.Hash = hash
		filename := ""
		if !util.CheckRequest(w, r) {
			return
		}
		if mode == "login" {
			filename = "login"
		} else if context.Login.Id == 0 {
			http.Redirect(w, r, "/manage/login", 303)
			return
		} else if mode == "logout" {
			account.Logout(r)
			http.SetCookie(w, &http.Cookie{
				Name:     "ww_tk",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   1,
				Value:    "logout",
			})
			http.Redirect(w, r, "/manage/", 303)
			return
		} else if mode == "" {
			filename = "manage"
		} else if mode == "newaccount" {
			filename = "newaccount"
		} else if mode == "category" {
			filename = "category"
			context.Mode = "登録"
			editid, err := strconv.Atoi(r.FormValue("edit"))
			if err == nil {
				cate, err := category.Get(editid)
				if err != nil {
					log.Println("ManageHandle category.Get")
					log.Println(err)
					Page500(w, err.Error())
					return
				}
				context.Mode = "編集"
				context.Category = cate
			}
		} else if mode == "categorylist" {
			filename = "categorylist"
			var err error
			context.CategoryList, err = category.All()
			if err != nil {
				log.Println("ManageHandle categorylist category.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else if mode == "question" {
			filename = "question"
			context.Mode = "入稿"
			editid, err := strconv.Atoi(r.FormValue("edit"))
			if err == nil {
				cue, err := question.Get(editid)
				if err != nil {
					log.Println("ManageHandle question.Get")
					log.Println(err)
					Page500(w, err.Error())
					return
				}
				context.Mode = "編集"
				context.Question = cue
			}
			context.CategoryList, err = category.All()
			if err != nil {
				log.Println("ManageHandle question category.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else if mode == "questionlist" {
			filename = "questionlist"
			var err error
			context.CategoryList, err = category.All()
			if err != nil {
				log.Println("ManageHandle questionlist category.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
			context.QuestionList, err = question.List()
			if err != nil {
				log.Println("ManageHandle questionlist question.List")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else if mode == "bloglist" {
			filename = "bloglist"
			var err error
			context.BlogList, err = blog.All()
			if err != nil {
				log.Println("ManageHandle bloglist blog.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else if mode == "blog" {
			filename = "blog"
			context.Mode = "登録"
			editid, err := strconv.Atoi(r.FormValue("edit"))
			if err == nil {
				context.Mode = "編集"
				context.Blog = blog.Get(editid)
			}
		} else if mode == "html" {
			filename = "html"
			var err error
			context.Pages, err = pagehtml.All()
			if err != nil {
				log.Println("ManageHandle html pagehtml.All")
				log.Println(err)
				Page500(w, err.Error())
				return
			}
		} else {
			Page404(w)
			return
		}
		if err := template.Must(template.ParseFiles("template/manage/"+filename+".html")).Execute(w, context); err != nil {
			log.Println(err)
			http.Error(w, "500", 500)
			return
		}
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func SocketHandle(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r2 *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	clients[ws] = r.URL.Path[len("/ws/"):]

	for {
		var msg SocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}
		msg.RoomId = r.URL.Path[len("/ws/"):]
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client, id := range clients {
			if id == msg.RoomId {
				if msg.Message != "" {
					err := client.WriteJSON(msg.Message)
					if err != nil {
						log.Printf("error: %v", err)
						client.Close()
						delete(clients, client)
					}
				} else {
					err := client.WriteJSON(msg)
					if err != nil {
						log.Printf("error: %v", err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}
	}
}

func ApiHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	mode := ""
	hash := ""
	if len(r.URL.Path) > len("/r/") {
		mode = r.URL.Path[len("/r/"):]
		if strings.Index(mode, "/") > 0 {
			hash = mode[strings.LastIndex(mode, "/")+1:]
			mode = mode[:strings.LastIndex(mode, "/")]
		}
	}

	if r.Method == http.MethodGet {
		if mode == "allquestions" {
			qs, err := question.All()
			if err != nil {
				log.Println("ApiHandle get allquestions question.All")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			bts, _ := json.Marshal(struct {
				ResultType int                 `json:"result_type"`
				Question   []question.Question `json:"question"`
			}{
				ResultType: 0,
				Question:   qs,
			})
			fmt.Fprintln(w, string(bts))
			return
		} else if mode == "categorysetting" {
			type Result struct {
				Category []category.Category `json:"Category"`
				Question []question.Question `json:"Question"`
			}
			ret := Result{}
			var err error
			ret.Category, err = category.All()
			if err != nil {
				log.Println("ApiHandle get categorysetting category.All")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ret.Question, err = question.All()
			if err != nil {
				log.Println("ApiHandle get categorysetting question.All")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			bts, _ := json.Marshal(ret)
			fmt.Fprint(w, string(bts))
			return
		} else if mode == "blog" {
			page, err := strconv.Atoi(r.FormValue("page"))
			if err != nil {
				ApiResponse(w, 200, -1, "insufficient parameter: page int")
				return
			}
			blogs := blog.Page(20, page)
			bts, _ := json.Marshal(ResultList{
				ResultType: 0,
				List:       blogs,
			})
			fmt.Fprint(w, string(bts))
			return
		} else if mode == "page_path" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			bts, _ := json.Marshal(struct {
				ResultType int               `json:"result_type"`
				PH         pagehtml.PageHtml `json:"ph"`
			}{
				ResultType: 0,
				PH:         pagehtml.Get(r.FormValue("page_path")),
			})
			fmt.Fprint(w, string(bts))
			return
		} else if mode == "vote" {
			voteend, err := member.VoteEnd(hash)
			if err != nil {
				log.Println("ApiHandle get vote member.VoteEnd")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ret := 0
			if voteend {
				ret = 1
			}
			ApiResponse(w, 200, ret, "")
			return
		}
		fmt.Fprintf(w, mode)
	} else if r.Method == http.MethodPost {
		r.ParseMultipartForm(32 << 20)
		if mode == "waiting" {
			myid, err := strconv.Atoi(r.FormValue("id"))
			if err != nil {
				ApiResponse(w, 400, -1, "insufficient parameter")
				return
			}
			db := database.Connect()
			defer db.Close()
			mems, err := member.Room(hash, db)
			if err != nil {
				log.Println("ApiHandle post share member.Room")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("ApiHansle post share room.Get")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			if rm.Url != "" {
				type Member2 struct {
					Name string `json:"name"`
					Id   int    `json:"id"`
				}
				mems2 := make([]Member2, 0)
				me_exist := false
				for _, m := range mems {
					if m.Id == myid {
						me_exist = true
					}
					mems2 = append(mems2, Member2{
						Name: m.Name,
						Id:   m.Id,
					})
				}
				if !me_exist {
					ApiResponse(w, 403, -1, "You are not includes to members.")
					return
				}
				bts, _ := json.Marshal(ResultList{
					ResultType: 0,
					Members:    mems2,
				})
				fmt.Fprint(w, string(bts))
			} else {
				ApiResponse(w, 404, -2, "room not found")
			}
			return
		} else if mode == "counting" {
			db := database.Connect()
			defer db.Close()
			if util.Isset(r, []string{"count"}) {
				voted, err := member.Voted(hash, db)
				if err != nil {
					log.Println("ApiHandle post counting member.Voted")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				bts, _ := json.Marshal(struct {
					ResultType int `json:"result_type"`
					Voted      int `json:"voted"`
				}{
					ResultType: 0,
					Voted:      voted,
				})
				fmt.Fprint(w, string(bts))
			} else if util.Isset(r, []string{"memid", "myid"}) {
				ck, err := r.Cookie("ww_voted")
				if err == nil {
					if ck.Value == "true" {
						ApiResponse(w, 200, 0, "already voted")
						return
					}
				}
				oppoId, err := strconv.Atoi(r.FormValue("memid"))
				if err != nil {
					ApiResponse(w, 400, -1, "Not selected opponent.")
					return
				}
				err = member.Vote(hash, oppoId, db)
				if err != nil {
					log.Println("ApiHandle post counting member.Vote")
					log.Println(err)
					ApiResponse(w, 400, -1, err.Error())
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "ww_voted",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   3600 * 24 * 3,
					Value:    "true",
				})
				ApiResponse(w, 200, 0, "voted")
			} else {
				ApiResponse(w, 400, -1, "Insufficient parameters")
			}
			return
		} else if mode == "addmember" {
			db := database.Connect()
			defer db.Close()
			if util.Isset(r, []string{"name", "url"}) {
				rm, err := room.Get(r.FormValue("url"), db)
				if err != nil {
					log.Println("ApiHandle post addmember room.Get")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				mems, err := member.Room(r.FormValue("url"), db)
				if err != nil {
					log.Println("ApiHandle post addmember member.Room")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				if len(mems) >= rm.PlayerCount {
					ApiResponse(w, 200, 1, "")
					return
				}
				newname := strings.TrimSpace(r.FormValue("name"))
				if newname == "" {
					ApiResponse(w, 400, 2, "")
					return
				}
				for _, m := range mems {
					if m.Name == newname {
						ApiResponse(w, 200, 2, "")
						return
					}
				}
				mem := member.Member{
					Name:    newname,
					RoomUrl: r.FormValue("url"),
				}
				newid, err := mem.Insert()
				if err != nil {
					log.Println("ApiHandle post addmember member.Insert")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				mem.Id = newid
				http.SetCookie(w, &http.Cookie{
					Name:     "ww_myname",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   3600 * 24 * 3,
					Value:    newname,
				})
				http.SetCookie(w, &http.Cookie{
					Name:     "ww_myid",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   3600 * 24 * 3,
					Value:    strconv.Itoa(mem.Id),
				})
				ApiResponse(w, 200, 0, strconv.Itoa(mem.Id))
				return
			} else {
				ApiResponse(w, 400, -1, "insufficient parameter")
				return
			}
		} else if mode == "delmember" {
			if util.Isset(r, []string{"id", "pass"}) {
				db := database.Connect()
				defer db.Close()
				mem := member.Get(hash, toInt(r.FormValue("id")), db)
				if mem.Id == 0 {
					ApiResponse(w, 400, -2, "")
					return
				}
				rm, err := room.Get(hash, db)
				if err != nil {
					log.Println("ApiHandle post delmember room.Get")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				if rm.Passcode == r.FormValue("pass") {
					err = mem.Delete(db)
					if err != nil {
						log.Println("ApiHandle post delmember mem.Delete")
						log.Print(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					ApiResponse(w, 200, 0, "")
				}
			} else {
				db := database.Connect()
				defer db.Close()
				err := member.CloseRoom(hash, db)
				if err != nil {
					log.Println("ApiHandle post delmember member.CloseRoom")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				ApiResponse(w, 200, 0, "")
			}
			return
		} else if mode == "odaireset" {
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("ApiHandle post odaireset room.Get")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			q, err := question.Search(rm.TalkCategory)
			if err != nil {
				log.Println("ApiHandle post odaireset question.Search")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			rand.Seed(time.Now().UnixNano())
			rnd := rand.Intn(len(q) - 1)
			// メンバーに割り振るお題を決めます。2進数で0ならval1、1ならval2をログイン順に割り振り
			omap := append(makeStr("0", rm.PlayerCount-rm.WordwolfCount), makeStr("1", rm.WordwolfCount)...)
			rand.Seed(time.Now().UnixNano() - 123456)
			rand.Shuffle(len(omap), func(i, j int) {
				omap[i], omap[j] = omap[j], omap[i]
			})
			omap_int, err := strconv.ParseInt(strings.Join(omap, ""), 2, 64)
			if err != nil {
				log.Println("ApiHandle post odaireset strconv.ParseInt")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			mems, err := member.Room(hash, db)
			if err != nil {
				log.Println("ApiHandle post odaireset member.Room")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			rm.TalkTheme = q[rnd].Id
			rm.OdaiMap = int(omap_int)
			err = rm.UpdateOdai(db)
			if err != nil {
				log.Println("ApiHandle post odaireset rm.UpdateOdai")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			for i, m := range mems {
				if omap[i] == "0" {
					m.Odai = q[rnd].Val1
				} else {
					m.Odai = q[rnd].Val2
				}
				err = m.UpdateOdai(db)
				if err != nil {
					log.Println("ApiHandle post odaireset m.UpdateOdai")
					log.Println(err)
					ApiResponse(w, 500, -1, err.Error())
					return
				}
			}
			ApiResponse(w, 200, 0, "")
			return
		} else if mode == "voteresult" {
			db := database.Connect()
			defer db.Close()
			rm, err := room.Get(hash, db)
			if err != nil {
				log.Println("ApiHandle post voteresult room.Get")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			if rm.Url == "" {
				ApiResponse(w, 404, -2, "")
				return
			}
			type Member2 struct {
				Name      string `json:"name"`
				VoteCount int    `json:"votecount"`
				Id        int    `json:"id"`
			}
			vt := make([]Member2, 0)
			ret := struct {
				ResultType int       `json:"result_type"`
				Vote       []Member2 `json:"vote"`
			}{
				ResultType: 0,
				Vote:       vt,
			}
			mems, err := member.Room(hash, db)
			if err != nil {
				log.Println("ApiHandle post voteresult member.Room")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			for _, m := range mems {
				ret.Vote = append(ret.Vote, Member2{
					Name:      m.Name,
					VoteCount: m.VoteCount,
					Id:        m.Id,
				})
			}
			bts, _ := json.Marshal(ret)
			fmt.Fprint(w, string(bts))
			return
		} else if mode == "announceresult" {
			type Announce struct {
				Name string `json:"name"`
				Odai string `json:"odai"`
				Id   int    `json:"id"`
			}
			ret := make([]Announce, 0)
			db := database.Connect()
			defer db.Close()
			mems, err := member.Room(hash, db)
			if err != nil {
				log.Println("ApiHandle post announceresult member.Room")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			for _, m := range mems {
				ret = append(ret, Announce{
					Name: m.Name,
					Odai: m.Odai,
					Id:   m.Id,
				})
			}
			bts, _ := json.Marshal(struct {
				ResultType int        `json:"result_type"`
				Announce   []Announce `json:"announce"`
			}{
				ResultType: 0,
				Announce:   ret,
			})
			fmt.Fprintln(w, string(bts))
			return
		} else if mode == "categorysetting" {
			type ImportData struct {
				Category []category.Category `json:"Category"`
				Question []question.Question `json:"Question"`
			}
			var data ImportData
			err := json.Unmarshal([]byte(r.FormValue("json")), &data)
			if err != nil {
				log.Println("ApiHandle post categorysetting json.Unmarshal")
				log.Println(err)
				ApiResponse(w, 400, -1, err.Error())
				return
			}
			db := database.Connect()
			defer db.Close()
			err = category.Insert(data.Category, db)
			if err != nil {
				log.Println("ApiHandle post categorysetting category.Insert")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			err = question.Insert(data.Question, db)
			if err != nil {
				log.Println("ApiHandle post categorysetting question.Insert")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ApiResponse(w, 200, 0, "")
			return
		} else if mode == "login" {
			if util.Isset(r, []string{"mail", "password"}) {
				a, err := account.Login(r.FormValue("mail"), r.FormValue("password"))
				if err != nil {
					ApiResponse(w, 400, -1, err.Error())
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "ww_tk",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   3600 * 24 * 3,
					Value:    a.Token,
				})
				ApiResponse(w, 200, 0, "ログインしました")
			} else {
				ApiResponse(w, 400, -1, "insufficient parameter: mail, password")
			}
			return
		} else if mode == "account" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			err := account.Account{
				Mail:     r.FormValue("mail"),
				Password: r.FormValue("password"),
			}.Insert()
			if err != nil {
				log.Println("ApiHandle post account account.Insert")
				log.Println(err)
				ApiResponse(w, 500, -1, "アカウントの作成に失敗しました")
				return
			}
			ApiResponse(w, 200, 0, "")
			return
		} else if mode == "category" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			if r.FormValue("id") != "" {
				//編集
				err := category.Category{
					Id:   toInt(r.FormValue("id")),
					Name: r.FormValue("name"),
				}.Update()
				if err != nil {
					if err.Error()[:1] == "." {
						ApiResponse(w, 200, -1, err.Error()[1:])
						return
					}
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				ApiResponse(w, 200, 1, "")
			} else {
				//登録
				err := category.Category{
					Name: r.FormValue("name"),
				}.Insert()
				if err != nil {
					if err.Error()[:1] == "." {
						ApiResponse(w, 200, -1, err.Error()[1:])
						return
					}
					ApiResponse(w, 500, -1, err.Error())
					return
				}
				ApiResponse(w, 200, 0, "")
			}
			return
		} else if mode == "question" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			if util.Isset(r, []string{"category", "val1", "val2"}) {
				if r.FormValue("id") != "" {
					//編集
					que, err := question.Get(toInt(r.FormValue("id")))
					if err != nil {
						log.Println("ApiHandle post question question.Get")
						log.Println(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					if que.Id == 0 {
						ApiResponse(w, 404, -1, "存在しないIDです")
						return
					}
					que.Category = toInt(r.FormValue("category"))
					que.Val1 = strings.TrimSpace(r.FormValue("val1"))
					que.Val2 = strings.TrimSpace(r.FormValue("val2"))
					err = que.Update()
					if err != nil {
						log.Println("ApiHandle post question que.Update")
						log.Println(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					ApiResponse(w, 200, 1, "")
				} else {
					//登録
					err := question.Question{
						Category: toInt(r.FormValue("category")),
						Val1:     strings.TrimSpace(r.FormValue("val1")),
						Val2:     strings.TrimSpace(r.FormValue("val2")),
					}.Insert()
					if err != nil {
						log.Println("ApiHandle post question question.Insert")
						log.Println(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					ApiResponse(w, 200, 0, "")
				}
			} else {
				ApiResponse(w, 400, -1, "insufficient parameter: category, val1, val2")
			}
			return
		} else if mode == "blog" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			if util.Isset(r, []string{"image", "title", "content"}) {
				db := database.Connect()
				defer db.Close()
				if r.FormValue("id") != "" {
					//編集
					b := blog.Get(toInt(r.FormValue("id")))
					if b.Id == 0 {
						ApiResponse(w, 404, -1, "存在しないIDです")
						return
					}
					b.Image = r.FormValue("image")
					b.Title = strings.TrimSpace(r.FormValue("title"))
					b.Content = template.HTML(r.FormValue("content"))
					err := b.Update()
					if err != nil {
						log.Println("ApiHandle post blog b.Update")
						log.Println(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					ApiResponse(w, 200, 1, "")
				} else {
					//登録
					err := blog.Blog{
						Image:   r.FormValue("image"),
						Title:   strings.TrimSpace(r.FormValue("title")),
						Content: template.HTML(r.FormValue("content")),
					}.Insert(db)
					if err != nil {
						log.Println("ApiHandle post blog blog.Insert")
						log.Println(err)
						ApiResponse(w, 500, -1, err.Error())
						return
					}
					ApiResponse(w, 200, 0, "")
				}
			} else {
				ApiResponse(w, 400, -1, "insufficient parameter: image, title, content")
			}
			return
		} else if mode == "html" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			if util.Isset(r, []string{"page_path", "left_html", "right_html", "top_html", "bottom_html"}) {
				err := pagehtml.PageHtml{
					PagePath: r.FormValue("page_path"),
				}.Insert()
				if err != nil {
					if err.Error()[:1] == "." {
						ApiResponse(w, 400, -1, err.Error()[1:])
						return
					}
					ApiResponse(w, 500, -1, err.Error())
				}
			} else {
				ApiResponse(w, 400, -1, "insufficient parameter: page_path, left_html, right_html, top_html, bottom_html")
			}
			return
		}
		fmt.Fprintf(w, mode)
	} else if r.Method == http.MethodPut {
		r.ParseMultipartForm(32 << 20)
		fmt.Fprintf(w, mode)
	} else if r.Method == http.MethodDelete {
		r.ParseMultipartForm(32 << 20)
		if mode == "category" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			id, err := strconv.Atoi(hash)
			if err != nil {
				ApiResponse(w, 404, -1, "id is not integer")
				return
			}
			err = category.Delete(id)
			if err != nil {
				log.Println("ApiHandle delete category category.Delete")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ApiResponse(w, 200, 0, "")
			return
		} else if mode == "question" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			id, err := strconv.Atoi(hash)
			if err != nil {
				ApiResponse(w, 404, -1, "id is not integer")
				return
			}
			err = question.Delete(id)
			if err != nil {
				log.Println("ApiHandle delete question question.Delete")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ApiResponse(w, 200, 0, "")
			return
		} else if mode == "blog" {
			if account.CheckLogin(r).Id == 0 {
				ApiResponse(w, 403, -1, "ログインしてください")
				return
			}
			id, err := strconv.Atoi(hash)
			if err != nil {
				ApiResponse(w, 404, -1, "id is not integer")
				return
			}
			err = blog.Delete(id)
			if err != nil {
				log.Println("ApiHandle delete blog blog.Delete")
				log.Println(err)
				ApiResponse(w, 500, -1, err.Error())
				return
			}
			ApiResponse(w, 200, 0, "")
			return
		}
		fmt.Fprintf(w, mode)
	} else {
		http.Error(w, "Method not allowed.", 405)
	}
}

func ApiResponse(w http.ResponseWriter, statuscode, result_type int, msg string) {
	bytes, err := json.Marshal(struct {
		ResultType int    `json:"result_type"`
		Message    string `json:"message"`
	}{
		ResultType: result_type,
		Message:    msg,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), statuscode)
		return
	}
	fmt.Fprintln(w, string(bytes))
}

func FaviconHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Add("Content-Type", "image/vnd.microsoft.icon")
		f, err := os.Open("static/favicon.ico")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer f.Close()
		io.Copy(w, f)
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func Page404(w http.ResponseWriter) {
	b, err := os.ReadFile("template/404.html")
	if err != nil {
		log.Print(err)
		b = []byte("404 Page Not Found")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	w.Write(b)
}

func Page500(w http.ResponseWriter, msg string) {
	b, err := ioutil.ReadFile("template/500.html")
	if err != nil {
		log.Print(err)
		b = []byte("500 Page Not Found")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(500)
	str := string(b)
	str = strings.Replace(str, "[message]", msg, -1)
	fmt.Fprintf(w, str)
}

func OutHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf8")
	b, err := os.ReadFile("nohup.out")
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf8")
		Page404(w)
		return
	}
	w.Write(b)
}

func SwjsHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf8")
	b, err := os.ReadFile("static/sw.js")
	if err != nil {
		w.Header().Set("Content-Type", "application/javascript; charset=utf8")
		Page404(w)
		return
	}
	w.Write(b)
}

func RobotsHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf8")
	b, err := os.ReadFile("static/robots.txt")
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		Page404(w)
		return
	}
	w.Write(b)
}

func SiteMapHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf8")
	b, err := os.ReadFile("static/sitemap.xml")
	if err != nil {
		w.Header().Set("Content-Type", "text/xml; charset=utf8")
		Page404(w)
		return
	}
	w.Write(b)
}
