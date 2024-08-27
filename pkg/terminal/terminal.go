package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"playlist-manager/internal/spotify"
	"playlist-manager/pkg/utils"

	spotifyapi "github.com/zmb3/spotify/v2"

	log "playlist-manager/pkg/logger"
)

func Display() (err error) {
	options := []string{"Autenticati su Spotify",
		"Visualizza le playlist del tuo account",
		"Visualizza i brani di una playlist del tuo account",
		"Salva una playlist (Backup)",
		"Carica una playlist (Restore)",
		"Esci"}

	var userID string

	//do while
	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("Playlist Manager v0.1.0\nSviluppato da Matteo Lombardi")
		fmt.Println("--------------------------------------------------------")
		fmt.Print("Autenticato: ")
		if spotify.IsAuthenticated() {
			fmt.Println("✅")

			//Display user ID
			if userID == "" {
				userID, err = spotify.GetUserID()
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
		fmt.Println("--------------------------------------------------------")
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
		case 1:
			err = spotify.Auth()
			if err != nil {
				return err
			}

		case 2:
			if !spotify.IsAuthenticated() {
				fmt.Println("Devi autenticarti prima di poter visualizzare le tue playlist")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}
			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}
			for i, p := range pl {
				fmt.Printf("%d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 3:
			if !spotify.IsAuthenticated() {
				fmt.Println("Devi autenticarti prima di poter visualizzare i brani di una playlist")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}
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

		case 4:
			if !spotify.IsAuthenticated() {
				fmt.Println("Devi autenticarti prima di poter salvare una playlist")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

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

			//Dir exists?
			if _, err := os.Stat("data/backup"); os.IsNotExist(err) {
				os.MkdirAll("data/backup", 0644) // Create dir if not exists
			}

			//Write file
			err = os.WriteFile("data/backup/"+string(pl[sel-1].ID)+".json", jsonData, 0644)
			if err != nil {
				return err
			}

			fmt.Println("Playlist salvata in data/backup/" + string(pl[sel-1].ID) + ".json")
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 5:
			if !spotify.IsAuthenticated() {
				fmt.Println("Devi autenticarti prima di poter caricare una playlist")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Get files
			files, err := os.ReadDir("data/backup")
			if err != nil {
				return err
			}
			for i, f := range files {
				fmt.Printf("%d. %s\n", i+1, f.Name())
			}

			//Select file
			fmt.Print("\nInserisci il numero del file da caricare: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel < 1 || sel > len(files) {
				fmt.Println("Selezione non valida")
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
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
			err = spotify.RestorePlaylist(playlist.TracksIDs, pl[sel-1].ID)
			if err != nil {
				return err
			}

			fmt.Println("Playlist caricata con successo")
			fmt.Printf("\nPremi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 6:
			fmt.Println("Esco...")
			return nil

		default:
			fmt.Println("Scelta non valida o non ancora implementata")
		}
		utils.ClearTerminal()
	}
}
