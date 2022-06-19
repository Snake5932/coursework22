package web_library

import (
	"DB_coursework/internal/config"
	"DB_coursework/internal/email_notifier"
	"DB_coursework/internal/models"
	server_gen "DB_coursework/internal/server"
	"DB_coursework/internal/utils"
	"DB_coursework/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type configuration struct {
	StaticDirPath   string
	FileServerPath  string
	ListenIPandPort string
	LogFile         string
	RedisIPandPort  string
	RedisPassword   string
	RedisPoolSize   int
	RedisTTL        int
	MailConfigPath  string
	DBUsername      string
	DBPassword      string
	DBIPandPort     string
	DBName          string
	PDFStorePath    string
	MaxUploadSize   int64
}

type Server struct {
	config   *configuration
	srv      *http.Server
	rdb      *redis.Client
	notifier *email_notifier.Notifier
	db       *pgxpool.Pool
	logger   *logger.Logger
	ctx      context.Context
}

func NewServer(configPath string) *Server {
	conf := configuration{}
	err := config.SetConfig(configPath, &conf)
	if err != nil {
		log.Fatal(err)
	}
	mailConf := email_notifier.Configuration{}
	err = config.SetConfig(conf.MailConfigPath, &mailConf)
	if err != nil {
		log.Fatal(err)
	}
	server := Server{}

	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHandler).Methods("GET")
	r.HandleFunc("/register", server.regHandler).Methods("GET")
	r.HandleFunc("/login", server.loginHandler).Methods("GET")
	r.HandleFunc("/account", server.mainPageHandler).Methods("GET")
	r.HandleFunc("/admin/account", server.mainAdminPageHandler).Methods("GET")
	r.HandleFunc("/admin/login", server.loginHandler).Methods("GET")

	s := r.PathPrefix("/api").Subrouter()
	s.HandleFunc("/approve", server.approve).Methods("POST")
	s.HandleFunc("/decline", server.decline).Methods("POST")
	s.HandleFunc("/getbookpage", server.getBookPage).Methods("POST")
	s.HandleFunc("/books", server.getBooks).Methods("POST")
	s.HandleFunc("/mybooks", server.getUserBooks).Methods("POST")
	s.HandleFunc("/ban", server.banUser).Methods("POST")
	s.HandleFunc("/getbookmeta", server.getBookMeta).Methods("POST")
	s.HandleFunc("/getrights", server.getAdminRights).Methods("POST")
	s.HandleFunc("/getlist", server.getList).Methods("POST")
	s.HandleFunc("/checklogin", server.checkLoginHandler).Methods("GET")
	s.HandleFunc("/sendmail", server.mailHandler).Methods("POST")
	s.HandleFunc("/checkmail", server.checkMailHandler).Methods("POST")
	s.HandleFunc("/exists", server.checkUniqueness).Methods("POST")
	s.HandleFunc("/register", server.registerHandler).Methods("POST")
	s.HandleFunc("/login", server.userLoginHandler).Methods("POST")
	s.HandleFunc("/logout", server.logoutHandler).Methods("POST")
	s.HandleFunc("/load", server.loadHandler).Methods("POST")
	s.HandleFunc("/addfile", server.addFileHandler).Methods("POST")
	s.HandleFunc("/change", server.changePassHandler).Methods("POST")

	server.config = &conf
	server.srv = &http.Server{
		Addr:              conf.ListenIPandPort,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		ReadTimeout:       0,
		IdleTimeout:       0,
		Handler:           r,
	}
	server.rdb = redis.NewClient(&redis.Options{
		Addr:            conf.RedisIPandPort,
		Password:        conf.RedisPassword,
		DB:              0,
		MaxRetries:      10,
		MinRetryBackoff: time.Second / 2,
		MaxRetryBackoff: time.Second / 2,
		PoolFIFO:        true,
		PoolSize:        conf.RedisPoolSize,
		PoolTimeout:     time.Second * 10,
	})
	server.notifier = &email_notifier.Notifier{Config: mailConf}
	server.logger = logger.Init(conf.LogFile)
	server.ctx = context.Background()
	server.db, err = pgxpool.Connect(server.ctx, "postgres://"+conf.DBUsername+":"+conf.DBPassword+"@"+conf.DBIPandPort+"/"+conf.DBName)
	if err != nil {
		return &Server{}
	}
	return &server
}

func (server *Server) getAdminRights(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin != "true" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	var rights pgtype.Bit
	err := server.db.QueryRow(server.ctx, "select rights from administrator where guid=$1", guid).Scan(&rights)
	if err != nil {
		server.logger.Error("can't get admin's rights for "+guid, err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	v, _ := rights.Value()
	w.Write([]byte(v.(string)))
}

func (server *Server) approve(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin != "true" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	book_id := r.FormValue("guid")

	tx, err := server.db.BeginTx(server.ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		server.logger.Error("can't start transaction for verification deletion", err)
		w.WriteHeader(500)
		return
	}
	defer tx.Rollback(server.ctx)
	_, err = tx.Exec(server.ctx, "delete from verification where book_id=$1 and admin_id=$2", book_id, guid)
	if err != nil {
		server.logger.Error("can't approve book", err)
		w.WriteHeader(500)
		return
	}
	_, err = tx.Exec(server.ctx, "REFRESH MATERIALIZED VIEW admin_book")
	if err != nil {
		server.logger.Error("can't refresh mat view after verification delete", err)
		w.WriteHeader(500)
		return
	}
	err = tx.Commit(server.ctx)
	if err != nil {
		server.logger.Error("can't commit transaction for verification deletion", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (server *Server) decline(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin != "true" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	book_id := r.FormValue("guid")
	username := r.FormValue("user")
	msg := r.FormValue("msg")
	if msg == "" {
		w.WriteHeader(400)
		return
	}
	tx, err := server.db.BeginTx(server.ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		server.logger.Error("can't start transaction for book deletion", err)
		w.WriteHeader(500)
		return
	}
	defer tx.Rollback(server.ctx)
	_, err = tx.Exec(server.ctx, "delete from book where guid=$1 and guid in "+
		"(select book_id from verification where admin_id=$2)", book_id, guid)
	if err != nil {
		server.logger.Error("can't decline book", err)
		w.WriteHeader(500)
		return
	}
	_, err = tx.Exec(server.ctx, "REFRESH MATERIALIZED VIEW admin_book")
	if err != nil {
		server.logger.Error("can't refresh mat view after book delete", err)
		w.WriteHeader(500)
		return
	}
	_, err = tx.Exec(server.ctx, "REFRESH MATERIALIZED VIEW user_book")
	if err != nil {
		server.logger.Error("can't refresh mat view after book delete", err)
		w.WriteHeader(500)
		return
	}
	err = tx.Commit(server.ctx)
	if err != nil {
		server.logger.Error("can't commit transaction after book delete", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	var email string
	err = server.db.QueryRow(server.ctx, "select email from end_user where nickname=$1", username).Scan(&email)
	if err != nil {
		server.logger.Error("can't send decline notification for "+username, err)
		return
	}
	var name string
	err = server.db.QueryRow(server.ctx, "select nickname from administrator where guid=$1", guid).Scan(&name)
	if err != nil {
		server.logger.Error("can't get admin's nickname for "+guid, err)
		w.WriteHeader(500)
		return
	}
	err = server.notifier.SendMail(email, "книга не прошла проверку", msg+"\nfrom "+name)
	if err != nil {
		server.logger.Error("can't send decline notification for "+username, err)
	}
}

func (server *Server) banUser(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin != "true" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	var rights pgtype.Bit
	var name string
	err := server.db.QueryRow(server.ctx, "select nickname, rights from administrator where guid=$1", guid).Scan(&name, &rights)
	if err != nil {
		server.logger.Error("can't get admin's rights for "+guid, err)
		w.WriteHeader(500)
		return
	}
	v, _ := rights.Value()
	if v.(string)[1] != '1' {
		w.WriteHeader(403)
		return
	}
	username := r.FormValue("user")
	msg := r.FormValue("msg")
	if msg == "" {
		w.WriteHeader(400)
		return
	}
	_, err = server.db.Exec(server.ctx, "update end_user set banned = not banned where nickname = $1", username)
	if err != nil {
		server.logger.Error("can't ban user "+username, err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	var email string
	err = server.db.QueryRow(server.ctx, "select email from end_user where nickname=$1", username).Scan(&email)
	if err != nil {
		server.logger.Error("can't send ban notification for "+username, err)
		return
	}
	err = server.notifier.SendMail(email, "бан", msg+"\nfrom "+name)
	if err != nil {
		server.logger.Error("can't send ban notification for "+username, err)
	}
}

func (server *Server) storePages(book_id, guid, status string, page int) {
	var rows pgx.Rows
	var err error
	if status == "admin" {
		rows, err = server.db.Query(server.ctx, "select page_data from page where num>=$1 and num <=$2 and book_id=$3 and"+
			" book_id in (select book_id from verification where admin_id=$4)", page, page+10, book_id, guid)
	} else if status == "user" {
		rows, err = server.db.Query(server.ctx, "select page_data from page where num>=$1 and num <=$2 and book_id=$3 and"+
			" book_id in (select guid from book where user_id=$4)", page, page+10, book_id, guid)
	} else {
		rows, err = server.db.Query(server.ctx, "select page_data from page where num>=$1 and num <=$2 and book_id=$3 and"+
			" book_id not in (select book_id from verification)", page, page+10, book_id)
	}
	if err != nil {
		server.logger.Error("can't retrieve pages to store", err)
		return
	}
	for rows.Next() {
		var bytea pgtype.Bytea
		err = rows.Scan(&bytea)
		if err != nil {
			server.logger.Error("error while iterating rows", err)
			page += 1
			continue
		}
		data, _ := bytea.Value()
		dataB := data.([]byte)
		key := book_id + "&" + guid + "&" + strconv.Itoa(page)
		_, err = server.rdb.Pipelined(server.ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(server.ctx, key,
				"blob", string(dataB),
				"closed", status,
				"guid", guid)
			pipe.Expire(server.ctx, key, 600*time.Second)
			return nil
		})
		if err != nil {
			server.logger.Error("can't store page data", err)
		}
		page += 1
	}
	if rows.Err() != nil {
		server.logger.Error("can't iterate pages to store", rows.Err())
	}
}

func (server *Server) getBookPage(w http.ResponseWriter, r *http.Request) {
	book_id := r.FormValue("guid")
	str_page := r.FormValue("page")
	page, err := strconv.Atoi(str_page)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	typ := r.FormValue("type")
	if typ == "gen" {
		data, err := server.rdb.HGetAll(server.ctx, book_id+"&&"+str_page).Result()
		if err != nil || len(data) == 0 {
			if err != nil {
				server.logger.Error("error while retrieving data from redis", err)
			}
			var bytea pgtype.Bytea
			err = server.db.QueryRow(server.ctx, "select page_data from page where num=$1 and book_id=$2 and"+
				" book_id not in (select book_id from verification)",
				page, book_id).Scan(&bytea)
			if err != nil {
				server.logger.Error("can't retrieve page data", err)
				w.WriteHeader(500)
				return
			}
			data, _ := bytea.Value()
			respB := data.([]byte)
			if len(respB) == 0 {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(200)
				w.Write(respB)
			}
			server.storePages(book_id, "", "open", page)
			return
		}
		if data["closed"] != "open" {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(200)
			w.Write([]byte(data["blob"]))
		}
	} else if typ == "user" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "false" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		data, err := server.rdb.HGetAll(server.ctx, book_id+"&"+guid+"&"+str_page).Result()
		if err != nil || len(data) == 0 {
			if err != nil {
				server.logger.Error("error while retrieving data from redis", err)
			}
			var bytea pgtype.Bytea
			err = server.db.QueryRow(server.ctx, "select page_data from page where num=$1 and book_id=$2 and"+
				" book_id in (select guid from book where user_id=$3)",
				page, book_id, guid).Scan(&bytea)
			if err != nil {
				server.logger.Error("can't retrieve page data", err)
				w.WriteHeader(500)
				return
			}
			data, _ := bytea.Value()
			respB := data.([]byte)
			if len(respB) == 0 {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(200)
				w.Write(respB)
			}
			server.storePages(book_id, guid, "user", page)
			return
		}
		if data["guid"] != guid {
			w.WriteHeader(403)
		} else {
			w.WriteHeader(200)
			w.Write([]byte(data["blob"]))
		}
	} else if typ == "admin" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "true" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		data, err := server.rdb.HGetAll(server.ctx, book_id+"&"+guid+"&"+str_page).Result()
		if err != nil || len(data) == 0 {
			if err != nil {
				server.logger.Error("error while retrieving data from redis", err)
			}
			var bytea pgtype.Bytea
			err = server.db.QueryRow(server.ctx, "select page_data from page where num=$1 and book_id=$2 and"+
				" book_id in (select book_id from verification where admin_id=$3)",
				page, book_id, guid).Scan(&bytea)
			if err != nil {
				server.logger.Error("can't retrieve page data", err)
				w.WriteHeader(500)
				return
			}
			data, _ := bytea.Value()
			respB := data.([]byte)
			if len(respB) == 0 {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(200)
				w.Write(respB)
			}
			server.storePages(book_id, guid, "admin", page)
			return
		}
		if data["guid"] != guid {
			w.WriteHeader(403)
		} else {
			w.WriteHeader(200)
			w.Write([]byte(data["blob"]))
		}
	} else {
		w.WriteHeader(500)
	}
}

func (server *Server) getBookMeta(w http.ResponseWriter, r *http.Request) {
	var respB []byte
	typ := r.FormValue("type")
	if typ == "check" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "true" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		book_id := r.FormValue("guid")
		var nickname, book_name, book_genre string
		var year, page_num int
		meta := models.BookMeta{}
		err := server.db.QueryRow(server.ctx, "select distinct nickname, book_name, book_genre, write_year, page_num from user_book where guid=$1 and"+
			" guid in (select book_id from verification where admin_id=$2)",
			book_id, guid).Scan(&nickname, &book_name, &book_genre, &year, &page_num)
		if err != nil {
			server.logger.Error("can't retrieve admin book meta data", err)
			w.WriteHeader(500)
			return
		}
		meta.Name = book_name
		meta.Nickname = nickname
		meta.Genre = utils.GetGenreCode(book_genre)
		meta.Year = year
		meta.Pagenum = page_num
		rows, err := server.db.Query(context.Background(), "select distinct author_name from user_book where guid=$1 and"+
			" guid in (select book_id from verification where admin_id=$2)", book_id, guid)
		if err != nil {
			server.logger.Error("can't get authors from admin meta", err)
			w.WriteHeader(500)
			return
		}
		for rows.Next() {
			var author string
			err = rows.Scan(&author)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			meta.Authors = append(meta.Authors, author)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate authors from admin meta", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(meta)
		if err != nil {
			server.logger.Error("can't marshal response for admin books list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "meta" {
		book_id := r.FormValue("guid")
		var nickname, book_name, book_genre string
		var year, page_num int
		meta := models.BookMeta{}
		err := server.db.QueryRow(server.ctx, "select distinct nickname, book_name, book_genre, write_year, page_num from user_book where guid=$1 and"+
			" guid not in (select book_id from verification)", book_id).Scan(&nickname, &book_name, &book_genre, &year, &page_num)
		if err != nil {
			server.logger.Error("can't retrieve book meta data", err)
			w.WriteHeader(500)
			return
		}
		meta.Name = book_name
		meta.Nickname = nickname
		meta.Genre = utils.GetGenreCode(book_genre)
		meta.Year = year
		meta.Pagenum = page_num
		rows, err := server.db.Query(context.Background(), "select distinct author_name from user_book where guid=$1 and"+
			" guid not in (select book_id from verification)", book_id)
		if err != nil {
			server.logger.Error("can't get authors from meta", err)
			w.WriteHeader(500)
			return
		}
		for rows.Next() {
			var author string
			err = rows.Scan(&author)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			meta.Authors = append(meta.Authors, author)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate authors from meta", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(meta)
		if err != nil {
			server.logger.Error("can't marshal response for books list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "self" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "false" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		book_id := r.FormValue("guid")
		var book_name, book_genre string
		var year, page_num int
		meta := models.BookMeta{}
		err := server.db.QueryRow(server.ctx, "select distinct book_name, book_genre, write_year, page_num from user_book where guid=$1 and user_id=$2",
			book_id, guid).Scan(&book_name, &book_genre, &year, &page_num)
		if err != nil {
			server.logger.Error("can't retrieve user book meta data", err)
			w.WriteHeader(500)
			return
		}
		meta.Name = book_name
		meta.Genre = utils.GetGenreCode(book_genre)
		meta.Year = year
		meta.Pagenum = page_num
		rows, err := server.db.Query(context.Background(), "select distinct author_name from user_book where guid=$1 and user_id=$2", book_id, guid)
		if err != nil {
			server.logger.Error("can't get authors from user meta", err)
			w.WriteHeader(500)
			return
		}
		for rows.Next() {
			var author string
			err = rows.Scan(&author)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			meta.Authors = append(meta.Authors, author)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate authors from user meta", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(meta)
		if err != nil {
			server.logger.Error("can't marshal response for user books list", err)
			w.WriteHeader(500)
			return
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(respB)
}

func (server *Server) getBooks(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		server.logger.Error("can't read book add request", err)
		w.WriteHeader(400)
		return
	}
	var bookData models.BookReq
	err = json.Unmarshal(body, &bookData)
	if err != nil {
		server.logger.Error("can't unmarshal book add request", err)
		w.WriteHeader(400)
		return
	}
	extraChecks := ""
	if bookData.Genre != "" {
		extraChecks += " and book_genre='" + utils.GetGenre(bookData.Genre) + "'"
	}
	minyear, res := utils.GetYear(bookData.MinYear)
	if res {
		extraChecks += " and write_year >= " + strconv.Itoa(minyear)
	}
	maxyear, res2 := utils.GetYear(bookData.MaxYear)
	if res2 {
		extraChecks += " and write_year <= " + strconv.Itoa(maxyear)
	}
	extraChecks += " and nickname ilike '%" + bookData.Nickname + "%' and book_name ilike '%" + bookData.Name + "%'"
	authorChecks := ""
	for _, author := range bookData.Authors {
		if author != "" {
			authorChecks += " or author_name ilike '%" + author + "%'"
		}
	}
	if authorChecks != "" {
		extraChecks += " and (" + authorChecks[4:] + ")"
	}
	rows, err := server.db.Query(context.Background(), "select distinct nickname, book_name, guid, book_genre, write_year from user_book where "+
		"guid not in (select book_id from verification)"+extraChecks+"order by guid limit 10 offset $1", strconv.Itoa(bookData.Page)+"0")
	if err != nil {
		server.logger.Error("can't get books gen", err)
		w.WriteHeader(500)
		return
	}
	books := struct {
		Books []models.BookResp
	}{Books: []models.BookResp{}}
	for rows.Next() {
		var nickname, name, book_guid, genre string
		var year int
		err = rows.Scan(&nickname, &name, &book_guid, &genre, &year)
		if err != nil {
			server.logger.Error("error while iterating rows", err)
			continue
		}

		authors := []string{}
		aut_rows, err := server.db.Query(context.Background(), "select distinct author_name from user_book where "+
			"guid not in (select book_id from verification) and guid=$1"+extraChecks, book_guid)
		if err != nil {
			server.logger.Error("can't get authors gen", err)
			w.WriteHeader(500)
			return
		}
		for aut_rows.Next() {
			var author string
			err = aut_rows.Scan(&author)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			authors = append(authors, author)
		}
		if aut_rows.Err() != nil {
			server.logger.Error("can't iterate authors gen", rows.Err())
			w.WriteHeader(500)
			return
		}

		books.Books = append(books.Books, models.BookResp{
			Name:     name,
			Nickname: nickname,
			Authors:  authors,
			Genre:    utils.GetGenreCode(genre),
			Year:     strconv.Itoa(year),
			Guid:     book_guid,
		})
	}
	if rows.Err() != nil {
		server.logger.Error("can't iterate books gen", rows.Err())
		w.WriteHeader(500)
		return
	}
	respB, err := json.Marshal(books)
	if err != nil {
		server.logger.Error("can't marshal response gen list", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(respB)
}

func (server *Server) getUserBooks(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin != "false" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		server.logger.Error("can't read book add request", err)
		w.WriteHeader(400)
		return
	}
	var bookData models.BookReq
	err = json.Unmarshal(body, &bookData)
	if err != nil {
		server.logger.Error("can't unmarshal book add request", err)
		w.WriteHeader(400)
		return
	}
	extraChecks := ""
	if bookData.Genre != "" {
		extraChecks += " and book_genre='" + utils.GetGenre(bookData.Genre) + "'"
	}
	minyear, res := utils.GetYear(bookData.MinYear)
	if res {
		extraChecks += " and write_year >= " + strconv.Itoa(minyear)
	}
	maxyear, res2 := utils.GetYear(bookData.MaxYear)
	if res2 {
		extraChecks += " and write_year <= " + strconv.Itoa(maxyear)
	}
	extraChecks += " and book_name ilike '%" + bookData.Name + "%'"
	authorChecks := ""
	for _, author := range bookData.Authors {
		if author != "" {
			authorChecks += " or author_name ilike '%" + author + "%'"
		}
	}
	if authorChecks != "" {
		extraChecks += " and (" + authorChecks[4:] + ")"
	}
	rows, err := server.db.Query(context.Background(), "select distinct book_name, guid, book_genre, write_year from user_book where "+
		"user_id=$1"+extraChecks+"order by guid limit 10 offset $2", guid, strconv.Itoa(bookData.Page)+"0")
	if err != nil {
		server.logger.Error("can't get books for user self", err)
		w.WriteHeader(500)
		return
	}
	books := struct {
		Books []models.BookResp
	}{Books: []models.BookResp{}}
	for rows.Next() {
		var name, book_guid, genre string
		var year int
		err = rows.Scan(&name, &book_guid, &genre, &year)
		if err != nil {
			server.logger.Error("error while iterating rows", err)
			continue
		}

		authors := []string{}
		aut_rows, err := server.db.Query(context.Background(), "select distinct author_name from user_book where "+
			"user_id=$1 and guid=$2"+extraChecks, guid, book_guid)
		if err != nil {
			server.logger.Error("can't get authors for user self", err)
			w.WriteHeader(500)
			return
		}
		for aut_rows.Next() {
			var author string
			err = aut_rows.Scan(&author)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			authors = append(authors, author)
		}
		if aut_rows.Err() != nil {
			server.logger.Error("can't iterate authors for user self", rows.Err())
			w.WriteHeader(500)
			return
		}

		books.Books = append(books.Books, models.BookResp{
			Name:     name,
			Nickname: "",
			Authors:  authors,
			Genre:    utils.GetGenreCode(genre),
			Year:     strconv.Itoa(year),
			Guid:     book_guid,
		})
	}
	if rows.Err() != nil {
		server.logger.Error("can't iterate books for user self", rows.Err())
		w.WriteHeader(500)
		return
	}
	respB, err := json.Marshal(books)
	if err != nil {
		server.logger.Error("can't marshal response for user self books list", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(respB)
}

func (server *Server) getList(w http.ResponseWriter, r *http.Request) {
	var respB []byte
	typ := r.FormValue("type")
	if typ == "user" {
		code, _, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, _ = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "true" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		user := r.FormValue("user")
		page := r.FormValue("page") + "0"
		rows, err := server.db.Query(context.Background(), "select nickname, email, banned from end_user where "+
			"nickname ilike '%"+user+"%' order by reg_date limit 10 offset $1", page)
		if err != nil {
			server.logger.Error("can't get users", err)
			w.WriteHeader(500)
			return
		}
		users := models.UserList{Users: []models.Userdata{}}
		i := 0
		for rows.Next() {
			var username, email string
			var banned bool
			err = rows.Scan(&username, &email, &banned)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			users.Users = append(users.Users, models.Userdata{Nickname: username, Email: email, Banned: banned, Num: i})
			i += 1
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate users", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(users)
		if err != nil {
			server.logger.Error("can't marshal response for users list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "author" {
		author := r.FormValue("author")
		rows, err := server.db.Query(context.Background(), "select author_name from author where "+
			"author_name ilike '%"+author+"%' order by author_name limit 30")
		if err != nil {
			server.logger.Error("can't get authors", err)
			w.WriteHeader(500)
			return
		}
		authors := struct {
			Authors []string
		}{Authors: []string{}}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			authors.Authors = append(authors.Authors, name)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate authors", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(authors)
		if err != nil {
			server.logger.Error("can't marshal response for authors list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "admin_book" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "true" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		page := r.FormValue("page") + "0"
		rows, err := server.db.Query(context.Background(), "select distinct book_name, guid from admin_book where "+
			"admin_id=$1 order by guid limit 10 offset $2", guid, page)
		if err != nil {
			server.logger.Error("can't get books for admin check", err)
			w.WriteHeader(500)
			return
		}
		books := struct {
			Books []struct {
				Name string
				Guid string
			}
		}{Books: []struct {
			Name string
			Guid string
		}{}}
		for rows.Next() {
			var name string
			var book_guid string
			err = rows.Scan(&name, &book_guid)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			books.Books = append(books.Books, struct {
				Name string
				Guid string
			}{name, book_guid})
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate books for admin", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(books)
		if err != nil {
			server.logger.Error("can't marshal response for admin books list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "book_name_gen" {
		name := r.FormValue("name")
		rows, err := server.db.Query(context.Background(), "select book_name from book where "+
			"guid not in (select book_id from verification) and book_name ilike '%"+name+"%' order by book_name limit 30")
		if err != nil {
			server.logger.Error("can't get book names gen", err)
			w.WriteHeader(500)
			return
		}
		names := struct {
			Entries []string
		}{Entries: []string{}}
		for rows.Next() {
			var book_name string
			err = rows.Scan(&book_name)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			names.Entries = append(names.Entries, book_name)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate book names gen", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(names)
		if err != nil {
			server.logger.Error("can't marshal response for names gen", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "book_name_user" {
		code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
		if code != 200 {
			isAdmin, code, guid = server_gen.LoginPass(server, r)
			if code != 200 {
				w.Header().Set("Location", "/login")
				w.WriteHeader(302)
				return
			}
		}
		if isAdmin != "false" {
			w.WriteHeader(403)
			return
		}
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		name := r.FormValue("name")
		rows, err := server.db.Query(context.Background(), "select book_name from book where "+
			"user_id=$1 and book_name ilike '%"+name+"%' order by book_name limit 30", guid)
		if err != nil {
			server.logger.Error("can't get book names for user", err)
			w.WriteHeader(500)
			return
		}
		names := struct {
			Entries []string
		}{Entries: []string{}}
		for rows.Next() {
			var book_name string
			err = rows.Scan(&book_name)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			names.Entries = append(names.Entries, book_name)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate book names for user", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(names)
		if err != nil {
			server.logger.Error("can't marshal response for names for user list", err)
			w.WriteHeader(500)
			return
		}
	} else if typ == "owner" {
		owner := r.FormValue("owner")
		rows, err := server.db.Query(context.Background(), "select nickname from end_user where "+
			"guid in (select distinct user_id from book where guid not in "+
			"(select book_id from verification)) and nickname ilike '%"+owner+"%' order by nickname limit 30")
		if err != nil {
			server.logger.Error("can't get owners", err)
			w.WriteHeader(500)
			return
		}
		owners := struct {
			Entries []string
		}{Entries: []string{}}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				server.logger.Error("error while iterating rows", err)
				continue
			}
			owners.Entries = append(owners.Entries, name)
		}
		if rows.Err() != nil {
			server.logger.Error("can't iterate owners", rows.Err())
			w.WriteHeader(500)
			return
		}
		respB, err = json.Marshal(owners)
		if err != nil {
			server.logger.Error("can't marshal response for owners list", err)
			w.WriteHeader(500)
			return
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(respB)
}

func (server *Server) addUser(nick, email, passHash string) error {
	_, err := server.db.Exec(server.ctx, "insert into end_user (nickname, email, pass_hash) values ($1, $2, $3)",
		nick, email, passHash)
	return err
}

func (server *Server) checkLoginHandler(w http.ResponseWriter, r *http.Request) {
	code, _, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, _ = server_gen.LoginPass(server, r)
		if code != 200 {
			w.WriteHeader(401)
			return
		}
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	w.WriteHeader(200)
	if isAdmin == "true" {
		w.Write([]byte("admin"))
	} else {
		w.Write([]byte("no"))
	}
}

func (server *Server) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	isAdmin, code, uuid := server_gen.LoginPass(server, r)
	if code != 200 {
		w.WriteHeader(code)
		return
	}
	cookie := utils.RandVerificationCode(10)
	server.SetCookie(cookie, uuid, isAdmin)
	http.SetCookie(w, &http.Cookie{
		Name:     "library-cookie",
		Value:    cookie,
		MaxAge:   7200,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	if isAdmin == "true" {
		w.Write([]byte("/admin/account"))
	} else {
		w.Write([]byte("/account"))
	}
}

func (server *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		server.logger.Error("Не удалось получить значения для регистрации", err)
		w.WriteHeader(400)
		return
	}
	email := r.FormValue("email")
	nickname := r.FormValue("nickname")
	password := r.FormValue("password")
	codeVal, err := server.rdb.HGetAll(server.ctx, email).Result()
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if codeVal["verified"] != "true" {
		w.WriteHeader(400)
		return
	}
	passHash := utils.GetPassHash(password)
	err = server.addUser(nickname, email, passHash)
	if err != nil {
		server.logger.Error("не удалось зарегистрировать пользователя", err)
		w.WriteHeader(400)
	} else {
		server.notifier.SendMail(email, "Регистрация", "Вы успешно зарегистрированы")
		cookie := utils.RandVerificationCode(10)
		var uuid string
		err = server.db.QueryRow(server.ctx, "select guid from end_user where nickname=$1", nickname).Scan(&uuid)
		if err == nil {
			server.SetCookie(cookie, uuid, "false")
			http.SetCookie(w, &http.Cookie{
				Name:     "library-cookie",
				Value:    cookie,
				MaxAge:   7200,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})
			w.Write([]byte("/account"))
		} else {
			w.WriteHeader(200)
		}
	}
}

func (server *Server) checkUniqueness(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		server.logger.Error("Не удалось получить значения для проверки уникальности", err)
		w.WriteHeader(400)
		return
	}
	val := r.FormValue("val")
	typ := r.FormValue("type")
	exists, err := utils.CheckExistence(server.db, "end_user", typ, val)
	if err != nil {
		server.logger.Error("не удалось проверить существование "+typ, err)
		w.WriteHeader(500)
		return
	}
	if exists {
		w.WriteHeader(400)
	} else {
		w.WriteHeader(200)
	}
}

func (server *Server) checkMailHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		server.logger.Error("Не удалось получить email и code", err)
		w.WriteHeader(400)
		return
	}
	email := r.FormValue("email")
	code := r.FormValue("code")
	codeVal, err := server.rdb.HGetAll(server.ctx, email).Result()
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if codeVal["code"] != code {
		w.WriteHeader(400)
		return
	}
	server.rdb.HSet(server.ctx, email, "verified", "true")
	server.rdb.Expire(server.ctx, email, time.Second*900)
	w.WriteHeader(200)
}

func (server *Server) mailHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		server.logger.Error("Не удалось получить email", err)
		w.WriteHeader(400)
		return
	}
	email := r.FormValue("email")
	codeVal := utils.RandVerificationCode(8)
	err = server.notifier.SendMail(email, "Код подтверждения", codeVal)
	if err != nil {
		server.logger.Error("Не удалось отправить email", err)
		w.WriteHeader(500)
		return
	}
	server.rdb.HSet(server.ctx, email, "code", codeVal, "verified", "false")
	w.WriteHeader(200)
}

func (server *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	var cookie, err = r.Cookie("library-cookie")
	if err != nil {
		w.WriteHeader(401)
	}
	cookieVal := cookie.Value
	server.rdb.Del(server.ctx, cookieVal)
	w.WriteHeader(200)
}

func (server *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/enter.html")
}

func (server *Server) regHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/register.html")
}

func (server *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	logged := false
	code, _, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code == 200 {
		logged = true
	} else {
		isAdmin, code, _ = server_gen.LoginPass(server, r)
		if code == 200 {
			logged = true
		}
	}
	if logged {
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
		if isAdmin == "true" {
			w.Header().Set("Location", "/admin/account")
		} else {
			w.Header().Set("Location", "/account")
		}
		w.WriteHeader(302)
		return
	}
	http.ServeFile(w, r, "web/login.html")
}

func (server *Server) mainPageHandler(w http.ResponseWriter, r *http.Request) {
	code, _, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, _ = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin == "true" {
		w.Header().Set("Location", "/login")
		w.WriteHeader(302)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	http.ServeFile(w, r, "web/mainPage.html")
}

func (server *Server) mainAdminPageHandler(w http.ResponseWriter, r *http.Request) {
	code, _, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, _ = server_gen.LoginPass(server, r)
		if code != 200 {
			w.Header().Set("Location", "/admin/login")
			w.WriteHeader(302)
			return
		}
	}
	if isAdmin == "false" {
		w.Header().Set("Location", "/admin/login")
		w.WriteHeader(302)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
	http.ServeFile(w, r, "web/mainAdminPage.html")
}

func (server *Server) changePassHandler(w http.ResponseWriter, r *http.Request) {
	code, uuid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, uuid = server_gen.LoginPass(server, r)
		if code != 200 {
			if isAdmin == "true" {
				w.Header().Set("Location", "/admin/login")
			} else {
				w.Header().Set("Location", "/login")
			}
			w.WriteHeader(302)
			return
		}
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}

	passHash := utils.GetPassHash(r.FormValue("password"))
	table := "end_user"
	if isAdmin == "true" {
		table = "administrator"
	}
	_, err := server.db.Exec(server.ctx, "update "+table+" set pass_hash = $1 where guid = $2", passHash, uuid)
	if err != nil {
		server.logger.Error("can't change password", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (server *Server) addFileHandler(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.WriteHeader(401)
			return
		}
	}
	if isAdmin == "true" {
		w.WriteHeader(401)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		server.logger.Error("can't read book add request", err)
		w.WriteHeader(400)
		return
	}
	var bookData = models.Book{}
	err = json.Unmarshal(body, &bookData)
	if err != nil {
		server.logger.Error("can't unmarshal book add request", err)
		w.WriteHeader(400)
		return
	}

	var book_guid string
	err = server.db.QueryRow(server.ctx, "select uuid_generate_v4()").Scan(&book_guid)
	fileDir := strings.TrimSuffix(bookData.File, filepath.Ext(bookData.File))
	items, err := ioutil.ReadDir(server.config.PDFStorePath + fileDir)
	if err != nil {
		server.logger.Error("can't get list of files for book", err)
		w.WriteHeader(500)
		return
	}
	tx, err := server.db.BeginTx(server.ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		server.logger.Error("can't start transaction", err)
		w.WriteHeader(500)
		return
	}
	defer tx.Rollback(server.ctx)

	year, res := utils.GetYear(bookData.Year)
	if res {
		_, err = tx.Exec(server.ctx, "insert into book(guid, user_id, book_name, book_genre, write_year) values($1, $2, $3, $4, $5)",
			book_guid, guid, bookData.Name, utils.GetGenre(bookData.Genre), year)
		if err != nil {
			server.logger.Error("can't add book", err)
			w.WriteHeader(500)
			return
		}
	} else {
		_, err = tx.Exec(server.ctx, "insert into book(guid, user_id, book_name, book_genre) values($1, $2, $3, $4)",
			book_guid, guid, bookData.Name, utils.GetGenre(bookData.Genre))
		if err != nil {
			server.logger.Error("can't add book", err)
			w.WriteHeader(500)
			return
		}
	}

	i := 0
	for _, item := range items {
		file, err := ioutil.ReadFile(server.config.PDFStorePath + fileDir + "/" + item.Name())
		if err != nil {
			server.logger.Error("can't read book page", err)
			w.WriteHeader(500)
			return
		}
		var bytea pgtype.Bytea
		err = bytea.Scan(file)
		if err != nil {
			server.logger.Error("can't scan bytea data", err)
			w.WriteHeader(500)
			return
		}
		_, err = tx.Exec(server.ctx, "insert into page(book_id, num, page_data) values($1, $2, $3)", book_guid, i, bytea)
		if err != nil {
			server.logger.Error("can't add data inside transaction", err)
			w.WriteHeader(500)
			return
		}
		i += 1
	}

	for _, author := range bookData.Authors {
		_, err = tx.Exec(server.ctx, "with aut as(insert into author(author_name) values($1) on conflict on constraint unique_author"+
			" do update set author_name=excluded.author_name returning guid) insert into aut_book_interm(author_id, book_id)"+
			" select guid, $2 from aut", author, book_guid)
		if err != nil {
			server.logger.Error("can't add data inside transaction", err)
			w.WriteHeader(500)
			return
		}
	}
	err = tx.Commit(server.ctx)
	if err != nil {
		server.logger.Error("can't commit transaction", err)
		w.WriteHeader(500)
		return
	}
	os.RemoveAll(server.config.PDFStorePath + fileDir)
	w.WriteHeader(200)
	_, err = server.db.Exec(server.ctx, "REFRESH MATERIALIZED VIEW user_book")
	if err != nil {
		server.logger.Error("can't refresh mat view after book addition", err)
	}
	_, err = server.db.Exec(server.ctx, "REFRESH MATERIALIZED VIEW admin_book")
	if err != nil {
		server.logger.Error("can't refresh mat view after book addition", err)
	}
}

func (server *Server) loadHandler(w http.ResponseWriter, r *http.Request) {
	code, guid, isAdmin, cookie := server_gen.LoginCookie(server, r)
	if code != 200 {
		isAdmin, code, guid = server_gen.LoginPass(server, r)
		if code != 200 {
			w.WriteHeader(401)
			return
		}
	}
	if isAdmin == "true" {
		w.WriteHeader(403)
		return
	}
	if cookie != nil {
		http.SetCookie(w, cookie)
	}

	var banned bool
	err := server.db.QueryRow(server.ctx, "select banned from end_user where guid=$1", guid).Scan(&banned)
	if err != nil {
		server.logger.Error("can't get ban status", err)
		w.WriteHeader(500)
		return
	}
	if banned {
		w.WriteHeader(403)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, server.config.MaxUploadSize)
	if err := r.ParseMultipartForm(server.config.MaxUploadSize); err != nil {
		server.logger.Error("file is too big", err)
		w.WriteHeader(400)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		server.logger.Error("can't read file", err)
		w.WriteHeader(400)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	filetype := http.DetectContentType(buff)
	if filetype != "application/pdf" {
		w.WriteHeader(400)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	fileDir := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	err = os.MkdirAll(server.config.PDFStorePath+fileDir+"/", 0777)
	if err != nil {
		server.logger.Error("can't make dirs", err)
		w.WriteHeader(500)
		return
	}

	dst, err := os.Create(fmt.Sprintf(server.config.PDFStorePath+"%s", fileHeader.Filename))
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var outb, errb bytes.Buffer
	err = utils.ExecCmdLine("qpdf --split-pages "+
		server.config.PDFStorePath+fileHeader.Filename+" "+
		server.config.PDFStorePath+fileDir+"/output-%d.pdf", &outb, &errb)
	if err != nil {
		server.logger.Error("can't split pdf"+"\n"+outb.String()+"\n"+errb.String(), err)
		w.WriteHeader(500)
		return
	}
	os.Remove(server.config.PDFStorePath + fileHeader.Filename)
	w.WriteHeader(200)
}

func (server *Server) QueryPass(table, login string, uuid, passFromBase *string) error {
	return server.db.QueryRow(server.ctx, "select guid, pass_hash from "+table+" where nickname=$1 or email=$1", login).Scan(uuid, passFromBase)
}

func (server *Server) GetCookieTTL(cookieVal string) (time.Duration, error) {
	return server.rdb.TTL(server.ctx, cookieVal).Result()
}

func (server *Server) GetUuid(cookieVal string) (map[string]string, error) {
	return server.rdb.HGetAll(server.ctx, cookieVal).Result()
}

func (server *Server) SetCookie(newCookie, uuid, isAdmin string) {
	server.rdb.HSet(server.ctx, newCookie, "uuid", uuid, "isAdmin", isAdmin)
	server.rdb.Expire(server.ctx, newCookie, 7200*time.Second)
}

func (server *Server) LogInfo(str string) {
	server.logger.Info(str)
}

func (server *Server) LogError(str string, err error) {
	server.logger.Error(str, err)
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.srv.Shutdown(ctx)
}

func (server *Server) GetIPAndPort() string {
	return server.config.ListenIPandPort
}

func (server *Server) ListenAndServe() error {
	return server.srv.ListenAndServe()
}

func (server *Server) CloseLog() {
	server.logger.Close()
}
