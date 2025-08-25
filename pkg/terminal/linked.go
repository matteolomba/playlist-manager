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
		fmt.Println("========================================================")
		fmt.Println("🎧 Playlist Manager " + VERSION + " 🎧\n✨ Sviluppato da " + termlink.ColorLink("Matteo Lombardi", "https://github.com/matteolomba", "italic yellow") + " ✨")
		fmt.Println("========================================================")
		displayAuthStatus()
		fmt.Println("========================================================")
		fmt.Println("🔗 -> Menù Playlist Collegate <- 🔗")
		fmt.Println("========================================================")
		fmt.Printf("🔙 0. Torna al menu principale\n")
		optionEmojis := []string{"👁️", "➕", "🗑️", "🔄"}
		for i, o := range options {
			fmt.Printf("%s %d. %s\n", optionEmojis[i], i+1, o)
		}
		fmt.Println("========================================================")
		fmt.Print("❓ Cosa vuoi fare? ")
		var sel int
		_, err := fmt.Scan(&sel)
		if err != nil {
			return err
		}

		fmt.Println()

		switch sel {
		case 0:
			fmt.Println("🔙 Ritorno al menu principale...")
			return nil
		case 1: // Show linked playlists
			utils.ClearTerminal()
			err = showLinkedPlaylists()
			if err != nil {
				return err
			}

		case 2: // Add linked playlist
			utils.ClearTerminal()
			err = addLinkedPlaylist()
			if err != nil {
				return err
			}

		case 3: // Remove linked playlist
			utils.ClearTerminal()
			err = removeLinkedPlaylist()
			if err != nil {
				return err
			}

		case 4: // Update songs in linked playlists
			utils.ClearTerminal()
			err = updateLinkedPlaylists()
			if err != nil {
				return err
			}

		default:
			fmt.Println("❌ Scelta non valida o non ancora implementata")
		}
		fmt.Printf("\n⏎ Premi invio per tornare al menu...")
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
		fmt.Println("===========================================")
		fmt.Println("🔗 -> Lista delle Playlist Collegate <- 🔗")
		fmt.Println("===========================================")
		fmt.Println()
		fmt.Println("🕵️ Nessuna playlist collegata, aggiungine una!")
	} else {
		fmt.Println("===========================================")
		fmt.Println("🔗 -> Lista delle Playlist Collegate <- 🔗")
		fmt.Println("===========================================")
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

			//Print playlist info in a formatted way
			fmt.Printf("\n🔗 %d. %s\n", i+1, tempPl.Name)
			fmt.Printf("   📄 File: %s\n", f.Name())

			// Print origin playlists
			fmt.Print("   📥 Origine: ")
			for j, origin := range tempPl.Origin {
				if j > 0 {
					fmt.Print(" ➕ ")
				}
				fmt.Printf("\"%s\"", origin.Name)
			}
			fmt.Println()

			// Print destination playlists
			fmt.Print("   🎯 Destinazione: ")
			for j, dest := range tempPl.Destination {
				if j > 0 {
					fmt.Print(" ➕ ")
				}
				fmt.Printf("\"%s\"", dest.Name)
			}
			fmt.Println()

			// Add separator line between playlists (except for the last one)
			if i < len(files)-1 {
				fmt.Println("   ─────────────────────────────────────")
			}
		}
	}
	return nil
}

func addLinkedPlaylist() (err error) {
	utils.ClearTerminal()
	fmt.Println("==============================================")
	fmt.Println("➕ -> Aggiunta Nuova Playlist Collegata <- ➕")
	fmt.Println("==============================================")
	fmt.Println()

	lp := linkedPlaylist{}
	fmt.Print("🎵 Che nome vuoi dare al collegamento tra le playlist? ")
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
		utils.ClearTerminal()
		fmt.Println("=============================================================")
		fmt.Println("🎯 -> Seleziona la playlist da aggiungere come origine <- 🎯")
		fmt.Println("=============================================================")
		fmt.Println()
		fmt.Println("🚪 0. Annulla e torna indietro")
		for i, p := range pl {
			fmt.Printf("📋 %d. %s - %s\n", i+1, p.Name, p.ID)
		}
		fmt.Println("=============================================================")
		fmt.Print("⏎ Inserisci il numero della playlist da aggiungere come origine: ")
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

	// Verifica se l'utente ha annullato senza configurare playlist origin
	if len(lp.Origin) == 0 {
		fmt.Println("🚪 Operazione annullata dall'utente")
		return nil
	}

	//Destination playlist selection
	for {
		utils.ClearTerminal()
		fmt.Println("==================================================================")
		fmt.Println("🎯 -> Seleziona la playlist da aggiungere come destinazione <- 🎯")
		fmt.Println("==================================================================")
		fmt.Println()
		fmt.Println("🚪 0. Annulla e torna indietro")
		for i, p := range pl {
			fmt.Printf("📋 %d. %s - %s\n", i+1, p.Name, p.ID)
		}
		fmt.Println("==================================================================")
		fmt.Print("⏎ Inserisci il numero della playlist da aggiungere come destinazione: ")
		var sel int
		_, err = fmt.Scan(&sel)
		if err != nil {
			return err
		}

		if sel == 0 {
			break
		} else if sel < 1 || sel > len(pl) {
			fmt.Println("❌ Selezione non valida")
		} else {
			lp.Destination = append(lp.Destination, Playlist{ID: string(pl[sel-1].ID), Name: pl[sel-1].Name})

			fmt.Print("❓ Vuoi aggiungere un'altra playlist come destinazione? (s/n) ")
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

	// Verifica se l'utente ha annullato senza configurare playlist destination
	if len(lp.Destination) == 0 {
		fmt.Println("🚪 Operazione annullata dall'utente")
		return nil
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
	utils.ClearTerminal()
	fmt.Println("=========================================")
	fmt.Println("🗑️ -> Rimozione Playlist Collegata <- 🗑️")
	fmt.Println("=========================================")
	fmt.Println()

	files, err := os.ReadDir("data/playlists")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("❌ Nessuna playlist collegata, aggiungine una!")
	} else {
		fmt.Println("=========================================================")
		fmt.Println("🎯 -> Seleziona la playlist collegata da rimuovere <- 🎯")
		fmt.Println("=========================================================")
		fmt.Println()
		fmt.Println("🚪 0. Annulla e torna indietro")
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

			//Print playlist info in a formatted way
			fmt.Printf("\n🔗 %d. %s\n", i+1, tempPl.Name)
			fmt.Printf("   📄 File: %s\n", f.Name())
			
			// Print origin playlists
			fmt.Print("   📥 Origine: ")
			for j, origin := range tempPl.Origin {
				if j > 0 {
					fmt.Print(" ➕ ")
				}
				fmt.Printf("🎵 %s", origin.Name)
			}
			fmt.Println()
			
			// Print destination playlists
			fmt.Print("   📤 Destinazione: ")
			for j, dest := range tempPl.Destination {
				if j > 0 {
					fmt.Print(" ➕ ")
				}
				fmt.Printf("🎯 %s", dest.Name)
			}
			fmt.Println()
		}

		fmt.Println("======================================================")
		fmt.Print("⏎ Inserisci il numero della playlist collegata da rimuovere: ")
		var sel int
		_, err = fmt.Scan(&sel)
		if err != nil {
			return err
		}

		if sel == 0 {
			fmt.Println("🚪 Operazione annullata dall'utente")
			return nil
		} else if sel < 1 || sel > len(files) {
			fmt.Println("❌ Selezione non valida")
			fmt.Printf("\n⏎ Premi invio per tornare al menu...")
			fmt.Scanf("\n\n")
		} else {
			err = os.Remove("data/playlists/" + files[sel-1].Name())
			if err != nil {
				return err
			}
			fmt.Println("✅ Playlist data/playlists/" + files[sel-1].Name() + " rimossa con successo")
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
	fmt.Println("=============================================")
	fmt.Println("🔄 -> Aggiornamento Playlist Collegate <- 🔄")
	fmt.Println("=============================================")
	fmt.Println()

	if len(files) == 0 {
		fmt.Println("❌ Nessuna playlist collegata, aggiungine una!")
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

	fmt.Println("=============================================================================")
	fmt.Println("🎵 ->          Aggiorno le canzoni nelle playlist collegate            <- 🎵")

	for _, pl := range playlists {
		fmt.Println("=============================================================================")
		fmt.Println("🎧 Playlist: " + pl.Name + " (" + pl.ID + ")")

		//Print linked playlist info (origin + destination)
		plString := pl.Origin[0].Name
		for i := 1; i < len(pl.Origin); i++ {
			plString += " + " + pl.Origin[i].Name
		}
		plString += " = " + pl.Destination[0].Name
		for i := 1; i < len(pl.Destination); i++ {
			plString += " + " + pl.Destination[i].Name
		}
		fmt.Println("🔗 " + plString)
		fmt.Println("⏳ ->                                                                    <- ⏳")

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

			//Get tracks that are only in the origin playlists (is the track in the destination playlist?)
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
				fmt.Println("❌ Nessuna canzone da aggiungere a " + p.Name)
			} else {
				if len(tracksToAdd) == 1 {
					fmt.Println("➕ Aggiungo 1 canzone a " + p.Name)
				} else {
					fmt.Printf("➕ Aggiungo %d canzoni a %s\n", len(tracksToAdd), p.Name)
				}
				err = spotify.AddTracksToPlaylist(tracksToAdd, spotifyapi.ID(p.ID))
				if err != nil {
					return err
				}
			}

			//Get tracks that are only in the destination playlists (is the track in the origin playlist?)
			var tracksToRemove []spotifyapi.ID
			for _, dt := range destTracks {
				found := false
				for _, t := range originTracks {
					if dt == t {
						found = true
						break
					}
				}
				if !found {
					tracksToRemove = append(tracksToRemove, dt)
				}
			}

			//Remove songs from destination playlists
			if len(tracksToRemove) == 0 {
				fmt.Println("❌ Nessuna canzone da rimuovere da " + p.Name)
			} else {
				if len(tracksToRemove) == 1 {
					fmt.Println("🗑️ Rimuovo 1 canzone da " + p.Name)
				} else {
					fmt.Printf("🗑️ Rimuovo %d canzoni da %s\n", len(tracksToRemove), p.Name)
				}
				err = spotify.RemoveTracksFromPlaylist(tracksToRemove, spotifyapi.ID(p.ID))
				if err != nil {
					return err
				}
			}
		}
	}

	fmt.Println("================================================================")
	fmt.Println("✅ Aggiornamento playlist collegate completato con successo! ✅")
	fmt.Println("================================================================")
	return nil
}
