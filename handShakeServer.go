package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var UsersNameId map[string]int   // Mapping from UserName to Id
var UsersIdSession map[int]int   // Mapping from Id to Session
var UsersIdReg map[int]string    // Mapping from Id to Rec_ID
var Sessions map[int]SessionData // Current running Sessions

var CurrentNumOfId int      // to assign new id for new users
var CurrentNumOfSession int // to assign new session

type SessionData struct {
	Graph    Graph
	Mappings map[string]string
}

type Graph struct {
	Nodes []Node
}

type Node struct {
	Data      string
	Id        int
	Neighbors []int
	Known     bool
	X         int
	Y         int
}

type GCMData struct {
	data             string
	registration_ids []string
}

func (n Node) String() string {
	return "{Id:" + fmt.Sprint(n.Id) + " Data: " + n.Data +
		" Neighbors: " + fmt.Sprint(n.Neighbors) + "}"
}

func UserExist(usrName string) bool {
	_, ok := UsersNameId[usrName]
	return ok
}

func main() {
	UsersIdSession = make(map[int]int)
	UsersNameId = make(map[string]int)
	Sessions = make(map[int]SessionData)
	UsersIdReg = make(map[int]string)

	CurrentNumOfSession = 2
	file, err := os.Open("data")
	if err != nil {
		log.Fatal(err.Error())
	}
	initialData := ParseData(file)
	for i := 1; i <= CurrentNumOfSession; i++ {
		Sessions[i] = initialData
	}

	http.HandleFunc("/SessionServer", LoginAndGetSession)
	http.HandleFunc("/GameServer", ConnectToSession)
	http.HandleFunc("/UpdateState", UpdateGameState)
	http.ListenAndServe(":33333", nil)

}

// GetUserSession expect parameter "user" and "session"
func LoginAndGetSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	regID := r.FormValue("regid")

	if !UserExist(usrName) {
		// User does not exist, Assign new id
		UsersNameId[usrName] = CurrentNumOfId + 1
		CurrentNumOfId++
	}

	UsersIdReg[UsersNameId[usrName]] = regID
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

//var UsersNameId map[string]int // Mapping from UserName to Id
//var UsersIdSession map[int]int // Mapping from Id to Session
//var Sessions map[int][]Node    // Current running Sessions
func ConnectToSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	usrSession, err := strconv.Atoi(r.FormValue("session"))
	if err != nil {
		log.Println(err)
	}

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

	jsonFormattedData, err := json.Marshal(Sessions[usrSession])
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("%s\n", jsonFormattedData)
	fmt.Fprintf(w, "%s", jsonFormattedData)
}

//regID, _ := strconv.Atoi(r.FormValue("regid"))
func ParseData(f *os.File) SessionData {
	// HERE IS WHERE WE PUT INITIAL GRAPH STATE INTO SESSION
	graphData := make([]Node, 0, 50)
	graphMapping := make(map[string]string)
	newId := 1
	scanner := bufio.NewScanner(f)
	set := make(map[string]Node)
	for scanner.Scan() {
		strsToks := strings.Split(scanner.Text(), ",")
		nodeName := strsToks[0]
		v, _ := strconv.Atoi(strsToks[1])
		visible := false
		if v == 1 {
			visible = true
		}
		x, _ := strconv.Atoi(strsToks[2])
		y, _ := strconv.Atoi(strsToks[3])
		fmt.Println(x)
		node, exists := set[nodeName]
		if !exists {
			node = Node{Data: nodeName, Id: newId, Known: visible, X: x, Y: y}
			graphMapping[strconv.Itoa(newId)] = nodeName
			set[nodeName] = node
			newId++
		} else {
			node.X = x
			node.Y = y
			node.Known = visible
		}
		scanner.Scan()
		strs := strings.Split(scanner.Text(), ",")
		neighbors := make([]int, 0)
		for _, edgeName := range strs {
			edge, exists := set[edgeName]
			if !exists {
				edge = (Node{Data: edgeName, Id: newId})
				graphMapping[strconv.Itoa(newId)] = edgeName
				set[edgeName] = edge
				newId++
			}
			neighbors = append(neighbors, edge.Id)
		}
		node.Neighbors = neighbors
		graphData = append(graphData, node)
	}

	return SessionData{Graph{graphData}, graphMapping}
}

// TAKES NAME/GAMEDATA
//type GCMData struct {
//data             map[string]string
//registration_ids []int
//}
func UpdateGameState(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	dataName := r.FormValue("data")
	usrId := UsersNameId[usrName]
	usrSess := UsersIdSession[usrId]
	fmt.Println(usrName)
	fmt.Println(dataName)

	UsersReg := []string{}
	for k, v := range UsersIdSession {
		if v == usrSess {
			UsersReg = append(UsersReg, UsersIdReg[k])
		}
	}
	fmt.Printf("%v\n", UsersReg)
	dataBodyGCM := GCMData{dataName, UsersReg}
	jsonGCMData, _ := json.Marshal(dataBodyGCM)
	fmt.Println(jsonGCMData)

	// Send it out
	gcm := &http.Client{}
	req, _ := http.NewRequest("POST", "https://android.googleapis.com/gcm/send", bytes.NewBuffer(jsonGCMData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key=AIzaSyCt7nNLPglsOiBoxCM5aSXbJw-93WkpMP4")

	resp, err := gcm.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
}
