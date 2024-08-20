package terminal

import (
	"fmt"
	"playlist-manager/internal/spotify"
	"playlist-manager/pkg/utils"

	log "playlist-manager/pkg/logger"
)

func Display() error {
	options := []string{"Autenticati su Spotify",
		"Visualizza le playlist del tuo account",
		"Visualizza i brani di una playlist del tuo account",
		"Salva una playlist (Backup)",
		"Carica una playlist (Restore)",
		"Esci"}

	//do while
	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("Playlist Manager v0.1.0\nSviluppato da Matteo Lombardi")
		fmt.Println("--------------------------------------------------------")
		fmt.Print("Autenticato:")
		if spotify.IsAuthenticated() {
			fmt.Println("✅")
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

		case 6:
			fmt.Println("Esco...")
			return nil

		default:
			fmt.Println("Scelta non valida o non ancora implementata")
		}
		utils.ClearTerminal()
	}
}
