package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"playlist-manager/internal/spotify"
	"playlist-manager/pkg/utils"

	"github.com/savioxavier/termlink"
	spotifyapi "github.com/zmb3/spotify/v2"

	log "playlist-manager/pkg/logger"
)

var userID string

const VERSION = "0.3.0"

func Display() (err error) {
	options := []string{
		"Visualizza le playlist del tuo account",
		"Visualizza i brani di una playlist del tuo account",
		"Salva una playlist (Backup)",
		"Carica una playlist (Restore)",
		"Visualizza e gestisci le playlist collegate",
	}

	err = spotify.Auth()
	if err != nil {
		return err
	}

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("Playlist Manager " + VERSION + "\nSviluppato da " + termlink.ColorLink("Matteo Lombardi", "https://github.com/matteolomba", "italic yellow"))
		fmt.Println("--------------------------------------------------------")
		displayAuthStatus(&userID)
		fmt.Println("--------------------------------------------------------")
		fmt.Println("-> Menù Principale <-")
		fmt.Println("--------------------------------------------------------")
		fmt.Printf("0. Esci\n")
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
			fmt.Println("Esco...")
			return nil

		case 1: // Get playlists
			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}
			for i, p := range pl {
				fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 2: // Get tracks from a playlist
			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}
			for i, p := range pl {
				fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Print("\nInserisci il numero della playlist di cui vuoi visualizzare i brani: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel < 1 || sel > len(pl) {
				fmt.Println("Selezione non valida")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}
			tracks, err := spotify.GetTracks(pl[sel-1].ID)
			if err != nil {
				return err
			}
			for i, t := range tracks {
				if t.Track.Track.ID == "" {
					log.Warn("Brano non disponibile, potrebbe essere un podcast o un brano non disponibile su Spotify")
				} else {
					text := fmt.Sprintf("%d. %s di ", i+1, t.Track.Track.Name)
					for _, a := range t.Track.Track.Artists {
						text += a.Name + ", "
					}
					text = text[:len(text)-2]
					fmt.Println(text)
				}
			}
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 3: // Save playlist (Backup) to JSON file
			//Get playlists
			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}
			for i, p := range pl {
				fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
			}

			//Select playlist
			fmt.Print("\nInserisci il numero della playlist di cui vuoi fare il backup: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel < 1 || sel > len(pl) {
				fmt.Println("Selezione non valida")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Get tracks and convert to JSON
			tracks, err := spotify.GetTracks(pl[sel-1].ID)
			if err != nil {
				return err
			}

			var tracksIDS []spotifyapi.ID
			for _, t := range tracks {
				if t.Track.Track.ID == "" {
					log.Warn("Brano non disponibile, potrebbe essere un podcast o un brano non disponibile su Spotify")
				} else {
					tracksIDS = append(tracksIDS, t.Track.Track.ID)
				}
			}
			playlist := spotify.Playlist{
				ID:        pl[sel-1].ID,
				Name:      pl[sel-1].Name,
				TracksIDs: tracksIDS,
			}
			jsonData, err := json.Marshal(playlist)
			if err != nil {
				return err
			}

			//Write file
			err = os.WriteFile("data/backup/"+string(pl[sel-1].ID)+".json", jsonData, 0644)
			if err != nil {
				return err
			}

			fmt.Println("Playlist salvata in data/backup/" + string(pl[sel-1].ID) + ".json")
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 4: // Restore playlist from JSON file

			//Get files
			files, err := os.ReadDir("data/backup")
			if err != nil {
				return err
			}
			fmt.Println("0. Torna al menu")
			for i, f := range files {
				//Read file
				tempData, err := os.ReadFile("data/backup/" + f.Name())
				if err != nil {
					return err
				}

				//Parse JSON
				var tempPl spotify.Playlist
				err = json.Unmarshal(tempData, &tempPl)
				if err != nil {
					return err
				}
				fmt.Printf("%d. %s (%s)\n", i+1, tempPl.Name, f.Name())
			}

			//Select file
			fmt.Print("\nInserisci il numero del file da caricare: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel < 0 || sel > len(files) {
				fmt.Println("Selezione non valida")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			} else {
				if sel == 0 {
					break
				}
			}

			//Read file
			data, err := os.ReadFile("data/backup/" + files[sel-1].Name())
			if err != nil {
				return err
			}

			//Parse JSON
			var playlist spotify.Playlist
			err = json.Unmarshal(data, &playlist)
			if err != nil {
				return err
			}

			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}
			for i, p := range pl {
				fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
			}

			fmt.Print("Inserisci il numero della playlist in cui caricare i brani: ")
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel < 1 || sel > len(pl) {
				fmt.Println("Selezione non valida")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Restore playlist
			err = spotify.AddTracksToPlaylist(playlist.TracksIDs, pl[sel-1].ID)
			if err != nil {
				return err
			}

			fmt.Println("Playlist caricata con successo")
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 5: // Manage linked playlists
			utils.ClearTerminal()
			err := linkedMenu()
			if err != nil {
				return err
			}

		default:
			fmt.Println("Scelta non valida o non ancora implementata")
		}
		utils.ClearTerminal()
	}
}

func displayAuthStatus(userID *string) {
	fmt.Print("Autenticato: ")
	if spotify.IsAuthenticated() {
		fmt.Println("✅")

		//Display user ID
		if *userID == "" {
			userID, err := spotify.GetUserID()
			if err != nil {
				log.Error("Errore nel recupero dell'ID dell'utente", "error", err)
			} else {
				fmt.Println("Utente:", userID)
			}
		} else {
			fmt.Println("Utente:", userID)
		}

	} else {
		fmt.Println("❌")
	}
}
