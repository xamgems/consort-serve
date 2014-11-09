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
	Width    float64
	Height   float64
	Mappings map[string]string
}

type Graph struct {
	Nodes []*Node
}

type Node struct {
	Data      string
	Id        int
	Neighbors []int
	Known     bool
	X         float64
	Y         float64
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
	file, err := os.Open("friends.txt")
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
	graphData := make([]*Node, 0, 50)
	graphMapping := make(map[string]string)
	newId := 1
	scanner := bufio.NewScanner(f)
	set := make(map[string]*Node)
	scanner.Scan()
	graphInfo := strings.Split(scanner.Text(), " ")
	graphWidth, _ := strconv.ParseFloat(graphInfo[2], 64)
	graphHeight, _ := strconv.ParseFloat(graphInfo[3], 64)
	for scanner.Scan() {
		strsToks := strings.Split(scanner.Text(), " ")
		lineType := strsToks[0]
		if lineType == "stop" {
			break
		}
		if lineType == "node" {
			visible := false
			if newId == 1 {
				visible = true
			}
			nodeName := strsToks[1]
			x, _ := strconv.ParseFloat(strsToks[2], 64)
			y, _ := strconv.ParseFloat(strsToks[3], 64)
			node := &(Node{Data: nodeName, Id: newId, Known: visible, X: x, Y: y})
			graphMapping[strconv.Itoa(newId)] = nodeName
			set[nodeName] = node
			newId++
			graphData = append(graphData, node)
		} else {
			start := strsToks[1]
			end := strsToks[2]
			startNode := set[start]
			endNode := set[end]
			startNode.Neighbors = append(startNode.Neighbors, endNode.Id)
			endNode.Neighbors = append(endNode.Neighbors, startNode.Id)
		}
	}

	return SessionData{Graph{graphData}, graphWidth, graphHeight, graphMapping}
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
