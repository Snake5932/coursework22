package admin_tool

import (
	"DB_coursework/internal/config"
	"DB_coursework/internal/utils"
	"context"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type configuration struct {
	PsqlIPandPort string
	DBName        string
}

type Tool struct {
	config *configuration
	db     *pgxpool.Pool
	ctx    context.Context
}

func NewTool(configPath, username, password string) (*Tool, error) {
	conf := configuration{}
	err := config.SetConfig(configPath, &conf)
	if err != nil {
		log.Fatal(err)
	}
	tool := Tool{}
	tool.config = &conf
	tool.ctx = context.Background()
	tool.db, err = pgxpool.Connect(tool.ctx, "postgres://"+username+":"+password+"@"+conf.PsqlIPandPort+"/"+conf.DBName)
	if err != nil {
		return &Tool{}, err
	}
	return &tool, nil
}

func (tool *Tool) CheckEmail(email string) (bool, error) {
	return utils.CheckExistence(tool.db, "administrator", "email", email)
}

func (tool *Tool) CheckNickName(nick string) (bool, error) {
	return utils.CheckExistence(tool.db, "administrator", "nickname", nick)
}

func (tool *Tool) AddAdmin(nick, email, passHash, rights string) error {
	var bit pgtype.Bit
	bit.Scan(rights)
	_, err := tool.db.Exec(tool.ctx, "insert into administrator (nickname, email, pass_hash, rights) values ($1, $2, $3, $4)",
		nick, email, passHash, bit)
	return err
}

func (tool *Tool) DelAdmin(tbd string) error {
	_, err := tool.db.Exec(tool.ctx, "delete from administrator where nickname=$1 or email=$1", tbd)
	return err
}

func (tool *Tool) CloseConn() {
	tool.db.Close()
}
