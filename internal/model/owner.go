package model

type UserId uint32

type User struct {
	Name string
	Id   UserId
}

type GroupId uint32

type Group struct {
	Name string
	Id   GroupId
}
