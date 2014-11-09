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

var CurrentNumOfId int      // to assign new id for new users
var CurrentNumOfSession int // to assign new session

type Node struct {
	Data      string
	Id        int
	Neighbors []int
}

func main() {
	UsersIdSession = make(map[int]int)
	UsersNameId = make(map[string]int)
	Sessions = make(map[int][]Node)

	CurrentNumOfSession = 2
	InitialData := ParseData()
	for i := 1; i <= CurrentNumOfSession; i++ {
		Sessions[i] = InitialData
	}

	http.HandleFunc("/SessionServer", LoginAndGetSession)
	http.HandleFunc("/GameServer", ConnectToSession)
	http.ListenAndServe(":33333", nil)

}

// GetUserSession expect parameter "user" and "session"
func LoginAndGetSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")

	if !UserExist(usrName) {
		// User does not exist, Assign new id
		UsersNameId[usrName] = CurrentNumOfId + 1
		CurrentNumOfId++
	}

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
	_, ok := UsersNameId[usrName]
	return ok
}

//var UsersNameId map[string]int // Mapping from UserName to Id
//var UsersIdSession map[int]int // Mapping from Id to Session
//var Sessions map[int][]Node    // Current running Sessions
func ConnectToSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("name")
	usrSession, _ := strconv.Atoi(r.FormValue("session"))

	/*
		if newGame {
			CurrentAvaiSession++
			usrSession = CurrentAvaiSession
		} else if _, ok := Sessions[usrSession]; !ok {
			//Error. (new game not clicked, passed session does not exist
		}
	*/

	if _, ok := Sessions[usrSession]; !ok {
		fmt.Fprintf(w, "SESSION PASSED IN DOES NOT EXIST\n")
	}
	usrID := UsersNameId[usrName]
	UsersIdSession[usrID] = usrSession

	jsonFormattedData, _ := json.Marshal(Sessions[usrSession])
	fmt.Printf("%s\n", jsonFormattedData)
	fmt.Fprintf(w, "%s", jsonFormattedData)
}

func ParseData() []Node {
	// HERE IS WHERE WE PUT INITIAL GRAPH STATE INTO SESSION
	TempData := Node{
		Data:      "cow",
		Id:        1,
		Neighbors: []int{},
	}

	GraphData := make([]Node, 50)

	GraphData = append(GraphData, TempData)
	return GraphData
}
