package main

type RoomStruct struct {
	Id  int `json:"id"`
	UserId  string `json:"id"  binding:"required"`
	RoomId  string `json:"account"`
}
