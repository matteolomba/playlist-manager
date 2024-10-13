package terminal

import (
	"fmt"
	"playlist-manager/pkg/utils"

	"github.com/savioxavier/termlink"
)

type Playlist struct {
	ID   string
	Name string
}

/*
Model that represents a linked playlist
A linked playlist contains:
- ID: the ID of the playlist (only for that program)
- Name: the name of the playlist (only for that program)
- Origin: the origin playlists (at least 2, where the songs will be taken from)
- Destination: the destination playlist/s (where the songs will be added from the origin playlists)
*/
type linkedPlaylist struct {
	ID          string
	Name        string
	Origin      []Playlist
	Destination []Playlist
}

func linkedMenu() (err error) {
	options := []string{"Visualizza le playlist collegate",
		"Aggiungi una playlist collegata",
		"Rimuovi una playlist collegata",
		"Aggiorna le canzoni nelle playlist collegate",
	}

	var userID string
	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("Playlist Manager v0.2.1\nSviluppato da " + termlink.ColorLink("Matteo Lombardi", "https://github.com/matteolomba", "italic yellow"))
		fmt.Println("--------------------------------------------------------")
		displayAuthStatus(&userID)
		fmt.Println("--------------------------------------------------------")
		fmt.Println("-> Menù Playlist Collegate <-")
		fmt.Println("--------------------------------------------------------")
		fmt.Printf("0. Torna al menu principale\n")
		for i, o := range options {
			fmt.Printf("%d. %s\n", i+1, o)
		}
		fmt.Println("--------------------------------------------------------")
		fmt.Print("Cosa vuoi fare? ")
		var sel int
		_, err := fmt.Scan(&sel)
		if err != nil {
			return err
		}

		fmt.Println()

		switch sel {
		case 0:
			return nil
		case 1: // Manage linked playlists

		/*case 1: // List linked playlists

		case 2: // Add linked playlist

		case 3: // Remove linked playlist

		case 4: // Update songs in linked playlists*/

		default:
			fmt.Println("Scelta non valida o non ancora implementata")
		}
		utils.ClearTerminal()
	}
}

func showLinkedPlaylists() (err error) {

	fmt.Printf("\nPremi invio per tornare al menu...")
	fmt.Scanf("\n\n")
	return nil
}

func addLinkedPlaylist() {
}

func removeLinkedPlaylist() {
}
