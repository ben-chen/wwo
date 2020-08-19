package main

import (
	"log"
	"net/http"
	"html/template"
	"os"
	"fmt"
	"math/rand"
	"time"
	"strconv"
)

type Player struct {
	Name string
	Role string
	Mayor bool
}

type Table struct {
	Players []*Player
}

type GameData struct {
	MyName string
	MyRole string
	AllNames []string
	AllRoles []string
	RoomID string
}

var tbl Table
var roles []string
var seer_roles []string
var revealed bool

func getNames(t *Table) []string {
	var l []string
	for _, p := range t.Players {
		n := p.Name
		if p.Mayor {
			n += "✦"
		}
		l = append(l, n)
	}
	return l
}

func isIn(p *Player, t *Table) bool {
	for _, v := range t.Players {
		if v.Name == p.Name {
			return true
		}
	}
	return false
}

func addPlayer(p *Player, t *Table) {
	if !isIn(p, t) {
		t.Players = append(t.Players, p)
	}
}

func getPlayer(name string) (*Player) {
	for _, p := range tbl.Players {
		if p.Name == name {
			return p
		}
	}
	null_player := Player{Name: "", Role: "", Mayor: false}
	return &null_player
}

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}

	t, err := template.ParseFiles(wd + "/html/" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		if revealed {
			http.Redirect(w, r, "/Reveal/", http.StatusSeeOther)
		}

		player_name := r.URL.Path[len("/Game/"):]
		if player_name == "" {
			renderTemplate(w, "register", GameData{MyName: "", MyRole: "", AllNames: getNames(&tbl), AllRoles: roles})
		} else {
			player := getPlayer(player_name)
			if player.Name == "" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			if player.Mayor {
				renderTemplate(w, "game_mayor", GameData{MyName: player.Name, MyRole: player.Role, AllNames: getNames(&tbl), AllRoles: roles})
			} else {
				renderTemplate(w, "game", GameData{MyName: player.Name, MyRole: player.Role, AllNames: getNames(&tbl), AllRoles: roles})
			}
		}


	case "POST":
		req := r.URL.Path[len("/Game/"):]

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		switch req[:4] {
		case "NewX":
			revealed = false
			http.Redirect(w, r, "/Game/" + req[4:], http.StatusSeeOther)
		case "Join":
			player := Player{Name: r.Form["player_name"][0], Role: "", Mayor: false}
			fmt.Printf("Adding %s\n", player.Name)
			addPlayer(&player, &tbl)
			http.Redirect(w, r, "/Game/" + player.Name, http.StatusSeeOther)
		case "Leav":
			name := r.URL.Path[len("/Game/Leave/"):]
			fmt.Printf("Removing %s\n", name)
			for j, v := range tbl.Players {
				if v.Name == name {
					tbl.Players = append(tbl.Players[:j],tbl.Players[j+1:]...)
					break;
				}
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		case "Assi":
			if len(tbl.Players) == 0 {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			player_name := req[len("Assign/"):]
			rand_perm := rand.Perm(len(tbl.Players))
			rand_num := rand.Intn(len(tbl.Players))
			for j, p := range tbl.Players {
				p.Role = roles[rand_perm[j]]
				p.Mayor = false
			}
			tbl.Players[rand_num].Mayor = true
			http.Redirect(w, r, "/Game/" + player_name, http.StatusSeeOther)
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/Game/", http.StatusSeeOther)
}

func roleHandler(w http.ResponseWriter, r *http.Request) {
	player_name := r.URL.Path[len("/RoleSelect/"):]
	player := getPlayer(player_name)
	renderTemplate(w, "roleselect", GameData{MyName: player.Name, MyRole: player.Role, AllNames: getNames(&tbl), AllRoles: roles})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	for i := 1; i <= 10; i++ {
		roles[i-1] = r.Form["role" + strconv.Itoa(i)][0]
	}
	http.Redirect(w, r, "/RoleSelect/" + r.URL.Path[len("/UpdateRoles/"):], http.StatusSeeOther)
}

func revealHandler(w http.ResponseWriter, r *http.Request) {
	player_name := r.URL.Path[len("/Reveal/"):]
	revealed = true

	var name_roles []string
	for _, p := range tbl.Players {
		mayor_tag := ""
		if p.Mayor {
			mayor_tag = " (Mayor)"
		}
		name_roles = append(name_roles, p.Name + mayor_tag + ": <" + p.Role + ">")
	}
	renderTemplate(w, "reveal", GameData{MyName: player_name, AllNames: name_roles})
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	tbl.Players = make([]*Player, 0, 9)
	roles = []string{"Villager", "Werewolf", "Seer", "Villager", "Villager", "Villager", "Mason", "Mason", "Fortune Teller", "Cow"}
	seer_roles = []string{"Werewolf", "Seer"}
	revealed = false
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/Game/", gameHandler)
	http.HandleFunc("/RoleSelect/", roleHandler)
	http.HandleFunc("/UpdateRoles/", updateHandler)
	http.HandleFunc("/Reveal/", revealHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
