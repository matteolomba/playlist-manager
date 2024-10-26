package terminal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"playlist-manager/internal/spotify"
	"playlist-manager/pkg/utils"

	"github.com/savioxavier/termlink"
	spotifyapi "github.com/zmb3/spotify/v2"
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
		case 1: // Show linked playlists
			err = showLinkedPlaylists()
			if err != nil {
				return err
			}

		case 2: // Add linked playlist
			err = addLinkedPlaylist()
			if err != nil {
				return err
			}

		case 3: // Remove linked playlist
			err = removeLinkedPlaylist()
			if err != nil {
				return err
			}

		case 4: // Update songs in linked playlists
			err = updateLinkedPlaylists()
			if err != nil {
				return err
			}

		default:
			fmt.Println("Scelta non valida o non ancora implementata")
		}
		fmt.Printf("\nPremi invio per tornare al menu...")
		fmt.Scanf("\n\n")
		utils.ClearTerminal()
	}
}

func showLinkedPlaylists() (err error) {
	//Get files
	files, err := os.ReadDir("data/playlists")
	if err != nil {
		return err
	}
	utils.ClearTerminal()

	if len(files) == 0 {
		fmt.Println("------------------------------------")
		fmt.Println("-> Lista delle Playlist Collegate <-")
		fmt.Println("------------------------------------")
		fmt.Println()
		fmt.Println("Nessuna playlist collegata, aggiungine una!")
	} else {
		fmt.Println("------------------------------------")
		fmt.Println("-> Lista delle Playlist Collegate <-")
		fmt.Println("------------------------------------")
		fmt.Println()
		for i, f := range files {
			//Read file
			tempData, err := os.ReadFile("data/playlists/" + f.Name())
			if err != nil {
				return err
			}

			//Parse JSON
			var tempPl linkedPlaylist
			err = json.Unmarshal(tempData, &tempPl)
			if err != nil {
				return err
			}

			//Print playlist info
			plString := tempPl.Origin[0].Name
			for i := 1; i < len(tempPl.Origin); i++ {
				plString += " + " + tempPl.Origin[i].Name
			}
			plString += " = " + tempPl.Destination[0].Name
			for i := 1; i < len(tempPl.Destination); i++ {
				plString += " + " + tempPl.Destination[i].Name
			}

			fmt.Printf("%d. %s (%s) -> %s\n", i+1, tempPl.Name, f.Name(), plString)
		}
	}
	return nil
}

func addLinkedPlaylist() (err error) {
	lp := linkedPlaylist{}
	fmt.Print("Che nome vuoi dare al collegamento tra le playlist? ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() != "" {
			lp.Name = scanner.Text()
			break
		}
	}

	pl, err := spotify.GetPlaylists()
	if err != nil {
		return err
	}

	//Origin playlist selection
	for {
		fmt.Println("------------------------------------------------------")
		fmt.Println("-> Seleziona la playlist da aggiungere come origine <-")
		fmt.Println("------------------------------------------------------")
		fmt.Println("0. Annulla e torna indietro")
		for i, p := range pl {
			fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
		}
		fmt.Println("------------------------------------------------------")
		fmt.Print("Inserisci il numero della playlist da aggiungere come origine: ")
		var sel int
		_, err = fmt.Scan(&sel)
		if err != nil {
			return err
		}

		if sel == 0 {
			break
		} else if sel < 1 || sel > len(pl) {
			fmt.Println("Selezione non valida")
		} else {
			lp.Origin = append(lp.Origin, Playlist{ID: string(pl[sel-1].ID), Name: pl[sel-1].Name})

			if len(lp.Origin) >= 2 {
				fmt.Print("Vuoi aggiungere un'altra playlist come origine? (s/n) ")
				var sel string
				_, err = fmt.Scan(&sel)
				if err != nil {
					return err
				}
				if sel != "s" {
					break
				}
			}
		}
	}

	//Destination playlist selection
	for {
		fmt.Println("-----------------------------------------------------------")
		fmt.Println("-> Seleziona la playlist da aggiungere come destinazione <-")
		fmt.Println("-----------------------------------------------------------")
		fmt.Println("0. Annulla e torna indietro")
		for i, p := range pl {
			fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
		}
		fmt.Println("-----------------------------------------------------------")
		fmt.Print("Inserisci il numero della playlist da aggiungere come destinazione: ")
		var sel int
		_, err = fmt.Scan(&sel)
		if err != nil {
			return err
		}

		if sel == 0 {
			break
		} else if sel < 1 || sel > len(pl) {
			fmt.Println("Selezione non valida")
		} else {
			lp.Destination = append(lp.Destination, Playlist{ID: string(pl[sel-1].ID), Name: pl[sel-1].Name})

			fmt.Print("Vuoi aggiungere un'altra playlist come destinazione? (s/n) ")
			var sel string
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel != "s" {
				break
			}
		}
	}

	//Generate the random ID
	lp.ID = utils.RandomString(10)

	jsonData, err := json.Marshal(lp)
	if err != nil {
		return err
	}

	//Write file
	err = os.WriteFile("data/playlists/"+lp.ID+".json", jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Playlist " + lp.Name + " salvata come data/playlists/" + lp.ID + ".json")
	return nil
}

func removeLinkedPlaylist() (err error) {
	files, err := os.ReadDir("data/playlists")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("Nessuna playlist collegata, aggiungine una!")
	} else {
		for i, f := range files {
			fmt.Println("--------------------------------------------------")
			fmt.Println("-> Seleziona la playlist collegata da rimuovere <-")
			fmt.Println("--------------------------------------------------")
			fmt.Println("0. Annulla e torna indietro")
			//Read file
			tempData, err := os.ReadFile("data/playlists/" + f.Name())
			if err != nil {
				return err
			}

			//Parse JSON
			var tempPl linkedPlaylist
			err = json.Unmarshal(tempData, &tempPl)
			if err != nil {
				return err
			}

			//Print playlist info
			plString := tempPl.Origin[0].Name
			for i := 1; i < len(tempPl.Origin); i++ {
				plString += " + " + tempPl.Origin[i].Name
			}
			plString += " = " + tempPl.Destination[0].Name
			for i := 1; i < len(tempPl.Destination); i++ {
				plString += " + " + tempPl.Destination[i].Name
			}

			fmt.Printf("%d. %s (%s) -> %s\n", i+1, tempPl.Name, f.Name(), plString)
		}

		fmt.Println("--------------------------------------------------")
		fmt.Print("Inserisci il numero della playlist collegata da rimuovere: ")
		var sel int
		_, err = fmt.Scan(&sel)
		if err != nil {
			return err
		}

		if sel == 0 {
			return nil
		} else if sel < 1 || sel > len(files) {
			fmt.Println("Selezione non valida")
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")
		} else {
			err = os.Remove("data/playlists/" + files[sel-1].Name())
			if err != nil {
				return err
			}
			fmt.Println("Playlist data/playlists/" + files[sel-1].Name() + " rimossa con successo")
		}

	}
	return nil
}

func updateLinkedPlaylists() (err error) {
	playlists := []linkedPlaylist{}

	//Get files
	files, err := os.ReadDir("data/playlists")
	if err != nil {
		return err
	}
	utils.ClearTerminal()

	if len(files) == 0 {
		fmt.Println("Nessuna playlist collegata, aggiungine una!")
		return nil
	} else {
		for _, f := range files {
			//Read file
			tempData, err := os.ReadFile("data/playlists/" + f.Name())
			if err != nil {
				return err
			}

			//Parse JSON
			var tempPl linkedPlaylist
			err = json.Unmarshal(tempData, &tempPl)
			if err != nil {
				return err
			}
			playlists = append(playlists, tempPl)
		}
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("-> Aggiorno le canzoni nelle playlist collegate <-")
	fmt.Println("--------------------------------------------------")

	for _, pl := range playlists {
		fmt.Println("Playlist: " + pl.Name)

		//-> Get tracks from origin playlists
		var originTracks []spotifyapi.ID
		for _, p := range pl.Origin {
			//Get tracks from origin playlist
			tracks, err := spotify.GetTrackIDs(spotifyapi.ID(p.ID))
			if err != nil {
				return err
			}
			originTracks = append(originTracks, tracks...)
		}

		//-> Compare tracks with destination playlists
		for _, p := range pl.Destination {
			//Get tracks from destination playlist
			destTracks, err := spotify.GetTrackIDs(spotifyapi.ID(p.ID))
			if err != nil {
				return err
			}

			//Remove tracks that are already in the destination playlist
			var tracksToAdd []spotifyapi.ID
			for _, t := range originTracks {
				found := false
				for _, dt := range destTracks {
					if t == dt {
						found = true
						break
					}
				}
				if !found {
					tracksToAdd = append(tracksToAdd, t)
				}
			}

			//Add songs to destination playlists
			if len(tracksToAdd) == 0 {
				fmt.Println("Nessuna canzone da aggiungere a " + p.Name)
				continue
			} else {
				fmt.Printf("Aggiungo %d canzoni a %s\n", len(tracksToAdd), p.Name)
				err = spotify.AddTracksToPlaylist(tracksToAdd, spotifyapi.ID(p.ID))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
