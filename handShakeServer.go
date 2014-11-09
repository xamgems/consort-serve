package main

import (
	"fmt"
	"net/http"
	//	"strconv"
)

var UsersNameId map[string]int // Mapping from UserName to Id
var UsersIdSession map[int]int // Mapping from Id to Session
var Sessions map[int]bool      // Current running Sessions

var CurrentAvaiId int // to assign new id for new users

func main() {
	UsersIdSession = make(map[int]int)
	UsersNameId = make(map[string]int)
	Sessions = make(map[int]bool)

	http.HandleFunc("/SessionServer", GetUserSession)
	http.HandleFunc("/GameServer", connectToSession)
	http.ListenAndServe(":33333", nil)

}

// GetUserSession expect parameter "user" and "session"
func GetUserSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")

	if !UserExist(usrName) {
		// User does not exist, Assign new id
		CurrentAvaiId++
		UsersNameId[usrName] = CurrentAvaiId
	}
	usrID := UsersNameId[usrName]

	usrSession := (usrID << 3 * 31) >> 2

	UsersIdSession[usrID] = usrSession
	Sessions[usrSession] = true

	SessionKeys := []int{}
	for k := range Sessions {
		SessionKeys = append(SessionKeys, k)
	}
	//fmt.Fprintf(w, "Connected successfully--\n user_name: %s\n user_id: %d\n user_session: %d", usrName, UsersNameId[usrName], UsersIdSession[UsersNameId[usrName]])
	fmt.Fprintf(w, "%v", SessionKeys)
}

func UserExist(usrName string) bool {
	if UsersNameId[usrName] == 0 {
		return false
	}
	return true
}

func connectToSession(w http.ResponseWriter, r *http.Request) {
	//usrID, _ := strconv.Atoi(r.FormValue("id"))
	//usrSession, _ := strconv.Atoi(r.FormValue("session"))
}
