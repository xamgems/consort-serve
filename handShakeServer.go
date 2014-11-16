package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/alexjlockwood/gcm"
	"github.com/mssola/user_agent"
)

var UsersNameId map[string]int   // Mapping from UserName to Id
var UsersIdSession map[int]int   // Mapping from Id to Session
var UsersIdReg map[int]string    // Mapping from Id to Rec_ID
var Sessions map[int]SessionData // Current running Sessions

var DataState map[int]StateMap // Mapping from GameState to Data

var CurrentGameState int    // Keep track of current game state
var CurrentNumOfId int      // to assign new id for new users
var CurrentNumOfSession int // to assign new session

type StateMap struct {
	State map[int]string
}

type SessionData struct {
	Graph    Graph
	Width    float64
	Height   float64
	Mappings map[string]string // Id to node name
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
	Data             string
	Registration_ids []string
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
	DataState = make(map[int]StateMap)

	CurrentGameState = 1
	CurrentNumOfSession = 2
	file, err := os.Open("data_set.plain")
	if err != nil {
		log.Fatal(err.Error())
	}
	initialData := ParseData(file)
	for i := 1; i <= CurrentNumOfSession; i++ {
		Sessions[i] = initialData
		dataState := StateMap{State: make(map[int]string)}
		DataState[i] = dataState
	}

	http.HandleFunc("/SessionServer", LoginAndGetSession)
	http.HandleFunc("/GameServer", ConnectToSession)
	http.HandleFunc("/UpdateState", UpdateGameState)

	// browser user call this function to ask for potential updates
	http.HandleFunc("/RequestUpdate", BrowserRequest)
	fmt.Println("Server is ready")
	http.ListenAndServe(":33333", nil)

}

// GetUserSession expect parameter "user" and "session"
func LoginAndGetSession(w http.ResponseWriter, r *http.Request) {
	ua := user_agent.New(r.UserAgent())
	usrName := r.FormValue("user")
	regID := "browser"
	// Commented this out for use
	 if ua.Mobile() {
		regID = r.FormValue("regid")
	 }

	fmt.Println("User: ", usrName, " RegId: ", regID)
	if !UserExist(usrName) {
		// User does not exist, Assign new id
		UsersNameId[usrName] = CurrentNumOfId
		CurrentNumOfId++
	}
	UsersIdReg[UsersNameId[usrName]] = regID

	SessionKeys := []int{}
	for k := range Sessions {
		SessionKeys = append(SessionKeys, k)
	}

	jsonFormatted, err := json.Marshal(SessionKeys)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("new user: %s connected\n", usrName)
	if ua.Mobile() {
		fmt.Printf("new user connectin with reg of %s\n" + regID)
	} else {
		fmt.Printf("Browser user connected.\n")
	}
	fmt.Fprintf(w, "%s\n", jsonFormatted)
}

func ConnectToSession(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	usrSession, err := strconv.Atoi(r.FormValue("session"))
	if err != nil {
		log.Println(err)
	}

	if _, ok := Sessions[usrSession]; !ok {
		fmt.Fprintf(w, "SESSION PASSED IN DOES NOT EXIST\n")
	}

	usrID := UsersNameId[usrName]
	UsersIdSession[usrID] = usrSession

	jsonFormattedData, err := json.Marshal(Sessions[usrSession])
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, "%s", jsonFormattedData)
}

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
			nodeName := strsToks[1]
			x, _ := strconv.ParseFloat(strsToks[2], 64)
			y, _ := strconv.ParseFloat(strsToks[3], 64)
			visible := false
			if strsToks[7] == "filled" {
				visible = true
			}
			node := &(Node{Data: strings.ToLower(nodeName),
				Id: newId, Known: visible, X: x, Y: y})
			graphMapping[strconv.Itoa(newId)] = strings.ToLower(nodeName)
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

// Expects Form data name "user" and "data"
// Receives the correct string and broadcast to the rest of the
// players.
// NOTE: THIS METHOD IS NOT COMPLETE. WE SHOULD UPDATE THE GAME STATE DATA
//       IN THE "SESSION" MAPPING
func UpdateGameState(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	dataName := r.FormValue("data")
	usrId := UsersNameId[usrName]
	usrSess := UsersIdSession[usrId]

	CurrentGameState++
	DataState[usrSess].State[CurrentGameState] = dataName

	fmt.Printf("User: %s is requesting data from session %d\n", usrName, usrSess)
	fmt.Println("\tData:", dataName)
	UsersReg := []string{}
	fmt.Println("UsersIdReg:", UsersIdReg)
	for user, sessId := range UsersIdSession {
		fmt.Println("SessId:", sessId, "user:", user)
		if sessId == usrSess {
			UsersReg = append(UsersReg, UsersIdReg[user])
		}
	}
	fmt.Println("Gcm to RegIds:", UsersReg)
	data := map[string]interface{}{"data": dataName}
	msg := gcm.NewMessage(data, UsersReg...)
	sender := &gcm.Sender{ApiKey: "AIzaSyCt7nNLPglsOiBoxCM5aSXbJw-93WkpMP4"}
	_, err := sender.Send(msg, 10)
	if err != nil {
		fmt.Println("Failed", err)
		return
	}
	fmt.Println("Gcm Sent Message!")
}

// For browser
func BrowserRequest(w http.ResponseWriter, r *http.Request) {
	usrName := r.FormValue("user")
	usrGameState, err := strconv.Atoi(r.FormValue("gamestate"))
	if err != nil {
		log.Println(err)
	}
	usrId := UsersNameId[usrName]
	usrSess := UsersIdSession[usrId]
	fmt.Printf("User: %s is requestin data\n", usrName)
	fmt.Printf("User requesting data from session: %d\n", usrSess)
	dataStateMap := DataState[usrSess]
	strToUpdate := []string{}

	if CurrentGameState != usrGameState {
		// User is out of date.
		for i := usrGameState + 1; i < len(dataStateMap.State); i++ {
			strToUpdate = append(strToUpdate, dataStateMap.State[i])
		}
	}
	jsonFormatted, err := json.Marshal(strToUpdate)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("sending user with string: %s\n", jsonFormatted)
	fmt.Fprintf(w, "%s\n", jsonFormatted)

}
