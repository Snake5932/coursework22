package models

type UserList struct {
	Users []Userdata
}

type Userdata struct {
	Nickname string
	Email    string
	Banned   bool
	Num      int
}

type Book struct {
	Name    string
	File    string
	Authors []string
	Genre   string
	Year    string
}

type BookReq struct {
	Name     string
	Nickname string
	Authors  []string
	Genre    string
	MinYear  string
	MaxYear  string
	Page     int
}

type BookResp struct {
	Name     string
	Nickname string
	Authors  []string
	Genre    string
	Year     string
	Guid     string
}

type BookMeta struct {
	Name     string
	Nickname string
	Authors  []string
	Genre    string
	Year     int
	Pagenum  int
}
