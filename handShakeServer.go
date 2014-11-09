package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var UsersNameId map[string]int // Mapping from UserName to Id
var UsersIdSession map[int]int // Mapping from Id to Session
var Sessions map[int][]Node    // Current running Sessions

var CurrentAvaiId int // to assign new id for new users

type Node struct {
	Data      string
	Id        int
	Neighbors []int
}

func main() {
	UsersIdSession = make(map[int]int)
	UsersNameId = make(map[string]int)
	Sessions = make(map[int][]Node)

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

	// HERE IS WHERE WE PUT INITIAL GRAPH STATE INTO SESSION
	TempData := Node{
		Data:      "cow",
		Id:        1,
		Neighbors: []int{},
	}

	UsersIdSession[usrID] = usrSession
	Sessions[usrSession] = append(Sessions[usrSession], TempData)

	SessionKeys := []int{}
	for k := range Sessions {
		SessionKeys = append(SessionKeys, k)
	}

	//var x = struct {
	//	Sessions []int
	//}{SessionKeys}

	//fmt.Fprintf(w, "Connected successfully--\n user_name: %s\n user_id: %d\n user_session: %d", usrName, UsersNameId[usrName], UsersIdSession[UsersNameId[usrName]])

	jsonFormatted, err := json.Marshal(SessionKeys)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%s\n", jsonFormatted)
	fmt.Fprintf(w, "%s\n", jsonFormatted)
}

func UserExist(usrName string) bool {
	if UsersNameId[usrName] == 0 {
		return false
	}
	return true
}

//var UsersNameId map[string]int // Mapping from UserName to Id
//var UsersIdSession map[int]int // Mapping from Id to Session
//var Sessions map[int][]Node    // Current running Sessions
func connectToSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("name")
	usrID := UsersNameId[usrName]
	usrSession, _ := strconv.Atoi(r.FormValue("session"))
	UsersIdSession[usrID] = usrSession

	jsonFormattedData, _ := json.Marshal(Sessions[usrSession])
	fmt.Printf("%s\n", jsonFormattedData)
	fmt.Fprintf(w, "%s", jsonFormattedData)
}
