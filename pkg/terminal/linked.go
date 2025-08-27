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

	log "playlist-manager/pkg/logger"
)

// Struttura per rappresentare i dettagli di una traccia
type TrackDetails struct {
	ID     string
	Name   string
	Artist string
}

// Funzione helper per recuperare i dettagli delle tracce
func getTrackDetails(trackIDs []spotifyapi.ID) ([]TrackDetails, error) {
	if len(trackIDs) == 0 {
		return []TrackDetails{}, nil
	}

	// Usa la nuova funzione dell'API Spotify
	spotifyTracks, err := spotify.GetTrackDetails(trackIDs)
	if err != nil {
		log.Warn("Errore nel recupero dettagli tracce dall'API Spotify", "error", err)
		// In caso di errore, restituisce i dettagli base con gli ID
		var details []TrackDetails
		for _, id := range trackIDs {
			details = append(details, TrackDetails{
				ID:     string(id),
				Name:   string(id),
				Artist: "Informazioni non disponibili",
			})
		}
		return details, nil
	}

	var details []TrackDetails
	for _, track := range spotifyTracks {
		if track == nil {
			continue
		}

		// Costruisci la stringa degli artisti
		var artists []string
		for _, artist := range track.Artists {
			artists = append(artists, artist.Name)
		}
		artistsStr := "Artista sconosciuto"
		if len(artists) > 0 {
			if len(artists) == 1 {
				artistsStr = artists[0]
			} else {
				// Per più artisti, usa il formato "Artista1, Artista2"
				artistsStr = artists[0]
				for i := 1; i < len(artists); i++ {
					artistsStr += ", " + artists[i]
				}
			}
		}

		details = append(details, TrackDetails{
			ID:     string(track.ID),
			Name:   track.Name + " di " + artistsStr,
			Artist: artistsStr,
		})
	}

	return details, nil
}

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
			fmt.Println("   📥 Origine:")
			for _, origin := range tempPl.Origin {
				fmt.Printf("      ↪ %s\n", origin.Name)
			}

			// Print destination playlists
			fmt.Println("   🎯 Destinazione:")
			for _, dest := range tempPl.Destination {
				fmt.Printf("      ↪ %s\n", dest.Name)
			}
			fmt.Println()

			// Add separator line between playlists (except for the last one)
			if i < len(files)-1 {
				fmt.Println("───────────────────────────────────────")
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
			fmt.Println("   📥 Origine:")
			for _, origin := range tempPl.Origin {
				fmt.Printf("      ↪ %s\n", origin.Name)
			}

			// Print destination playlists
			fmt.Println("   🎯 Destinazione:")
			for _, dest := range tempPl.Destination {
				fmt.Printf("      ↪ %s\n", dest.Name)
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
	log.Info("Inizio aggiornamento playlist collegate")

	playlists := []linkedPlaylist{}

	//Get files
	files, err := os.ReadDir("data/playlists")
	if err != nil {
		log.Error("Errore lettura directory playlist", "error", err)
		return err
	}
	log.Info("Playlist collegate trovate", "count", len(files))

	utils.ClearTerminal()
	fmt.Println("=============================================")
	fmt.Println("🔄 -> Aggiornamento Playlist Collegate <- 🔄")
	fmt.Println("=============================================")
	fmt.Println()

	if len(files) == 0 {
		fmt.Println("❌ Nessuna playlist collegata, aggiungine una!")
		log.Warn("Nessuna playlist collegata trovata")
		return nil
	}

	// Menu per scegliere il tipo di operazione
	fmt.Println("🎯 Che tipo di aggiornamento vuoi eseguire?")
	fmt.Println("==========================================")
	fmt.Println("➕ 1. Solo aggiunta canzoni (da origine a destinazione)")
	fmt.Println("🗑️ 2. Solo rimozione canzoni (dalle destinazioni)")
	fmt.Println("🔄 3. Sia aggiungere che rimuovere canzoni")
	fmt.Println("🚪 0. Annulla e torna indietro")
	fmt.Println("==========================================")
	fmt.Print("❓ Cosa vuoi fare? ")

	var operationType int
	_, err = fmt.Scan(&operationType)
	if err != nil {
		return err
	}

	if operationType == 0 {
		fmt.Println("🚪 Operazione annullata dall'utente")
		return nil
	}

	if operationType < 1 || operationType > 3 {
		fmt.Println("❌ Scelta non valida")
		return nil
	}

	addSongs := operationType == 1 || operationType == 3
	removeSongs := operationType == 2 || operationType == 3

	log.Info("Tipo di operazione selezionata", "operationType", operationType, "addSongs", addSongs, "removeSongs", removeSongs)

	utils.ClearTerminal()
	fmt.Println("=============================================")
	fmt.Println("🔄 -> Aggiornamento Playlist Collegate <- 🔄")
	fmt.Println("=============================================")
	fmt.Println()

	for _, f := range files {
		log.Info("Caricamento playlist collegata", "file", f.Name())
		//Read file
		tempData, err := os.ReadFile("data/playlists/" + f.Name())
		if err != nil {
			log.Error("Errore lettura file playlist", "file", f.Name(), "error", err)
			return err
		}

		//Parse JSON
		var tempPl linkedPlaylist
		err = json.Unmarshal(tempData, &tempPl)
		if err != nil {
			log.Error("Errore parsing JSON playlist", "file", f.Name(), "error", err)
			return err
		}
		log.Info("Playlist collegata caricata", "name", tempPl.Name, "id", tempPl.ID, "origins", len(tempPl.Origin), "destinations", len(tempPl.Destination))
		playlists = append(playlists, tempPl)
	}

	for _, pl := range playlists {
		fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("│ 🎧 Playlist: %s (%s)\n", pl.Name, pl.ID)

		//Print linked playlist info (origin + destination)
		plString := "\"" + pl.Origin[0].Name + "\""
		for i := 1; i < len(pl.Origin); i++ {
			plString += " + \"" + pl.Origin[i].Name + "\""
		}
		plString += "  ➜  \"" + pl.Destination[0].Name + "\""
		for i := 1; i < len(pl.Destination); i++ {
			plString += " + \"" + pl.Destination[i].Name + "\""
		}
		fmt.Printf("│ 🔗 %s\n", plString)
		fmt.Println("│")
		fmt.Printf("│ ⏳ Elaborazione in corso...\n")
		fmt.Println("├──────────────────────────────────────────────────────────────────────────────────────────")

		//-> Get tracks from origin playlists
		var originTracks []spotifyapi.ID
		log.Info("Inizio recupero tracce da playlist origine", "linkedPlaylistName", pl.Name, "originCount", len(pl.Origin))

		for _, p := range pl.Origin {
			log.Info("Recupero tracce da playlist origine", "playlistName", p.Name, "playlistID", p.ID)
			//Get tracks from origin playlist
			tracks, err := spotify.GetTrackIDs(spotifyapi.ID(p.ID))
			if err != nil {
				log.Error("Errore nel recupero tracce da playlist origine", "playlistName", p.Name, "playlistID", p.ID, "error", err)
				return err
			}
			log.Info("Tracce recuperate da playlist origine", "playlistName", p.Name, "trackCount", len(tracks))
			originTracks = append(originTracks, tracks...)
		}
		log.Info("Totale tracce origine recuperate", "totalTracks", len(originTracks))

		//-> Compare tracks with destination playlists
		for _, p := range pl.Destination {
			log.Info("Inizio processamento playlist destinazione", "playlistName", p.Name, "playlistID", p.ID)
			//Get tracks from destination playlist
			destTracks, err := spotify.GetTrackIDs(spotifyapi.ID(p.ID))
			if err != nil {
				log.Error("Errore nel recupero tracce da playlist destinazione", "playlistName", p.Name, "playlistID", p.ID, "error", err)
				return err
			}
			log.Info("Tracce recuperate da playlist destinazione", "playlistName", p.Name, "trackCount", len(destTracks))

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
			log.Info("Tracce da aggiungere identificate", "playlistName", p.Name, "tracksToAddCount", len(tracksToAdd))

			//Add songs to destination playlists
			if addSongs {
				if len(tracksToAdd) == 0 {
					fmt.Printf("│ ❌ Nessuna canzone da aggiungere a %s\n", p.Name)
					log.Info("Nessuna traccia da aggiungere", "playlistName", p.Name)
				} else {
					// Mostra l'operazione in corso
					if len(tracksToAdd) == 1 {
						fmt.Printf("│ ⏳ Aggiunta di 1 canzone a %s...\n", p.Name)
					} else {
						fmt.Printf("│ ⏳ Aggiunta di %d canzoni a %s...\n", len(tracksToAdd), p.Name)
					}

					log.Info("Inizio aggiunta tracce alla playlist", "playlistName", p.Name, "playlistID", p.ID, "tracksCount", len(tracksToAdd))
					err = spotify.AddTracksToPlaylist(tracksToAdd, spotifyapi.ID(p.ID))
					if err != nil {
						log.Error("ERRORE nell'aggiunta tracce alla playlist", "playlistName", p.Name, "playlistID", p.ID, "error", err, "tracksCount", len(tracksToAdd))
						return err
					}
					log.Info("Tracce aggiunte con successo", "playlistName", p.Name, "tracksCount", len(tracksToAdd))

					// Mostra il risultato completato
					if len(tracksToAdd) == 1 {
						fmt.Printf("│ ✅ Aggiunta 1 canzone a %s\n", p.Name)
					} else {
						fmt.Printf("│ ✅ Aggiunte %d canzoni a %s\n", len(tracksToAdd), p.Name)
					}

					// Mostra l'elenco delle canzoni aggiunte
					fmt.Printf("│     🎵 Canzoni aggiunte:\n")
					trackDetails, err := getTrackDetails(tracksToAdd)
					if err != nil {
						log.Warn("Errore nel recupero dettagli tracce aggiunte", "error", err)
					} else {
						for _, track := range trackDetails {
							fmt.Printf("│       ↪ %s\n", track.Name)
						}
					}
				}
			} else {
				log.Info("Aggiunta canzoni saltata per scelta utente", "playlistName", p.Name)
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
			log.Info("Tracce da rimuovere identificate", "playlistName", p.Name, "tracksToRemoveCount", len(tracksToRemove))

			//Remove songs from destination playlists
			if removeSongs {
				if len(tracksToRemove) == 0 {
					fmt.Printf("│ ❌ Nessuna canzone da rimuovere da %s\n", p.Name)
					log.Info("Nessuna traccia da rimuovere", "playlistName", p.Name)
				} else {
					// Mostra l'operazione in corso
					if len(tracksToRemove) == 1 {
						fmt.Printf("│ ⏳ Rimozione di 1 canzone da %s...\n", p.Name)
					} else {
						fmt.Printf("│ ⏳ Rimozione di %d canzoni da %s...\n", len(tracksToRemove), p.Name)
					}

					log.Info("Inizio rimozione tracce dalla playlist", "playlistName", p.Name, "playlistID", p.ID, "tracksCount", len(tracksToRemove))
					err = spotify.RemoveTracksFromPlaylist(tracksToRemove, spotifyapi.ID(p.ID))
					if err != nil {
						log.Error("ERRORE nella rimozione tracce dalla playlist", "playlistName", p.Name, "playlistID", p.ID, "error", err, "tracksCount", len(tracksToRemove))
						return err
					}
					log.Info("Tracce rimosse con successo", "playlistName", p.Name, "tracksCount", len(tracksToRemove))

					// Mostra il risultato completato
					if len(tracksToRemove) == 1 {
						fmt.Printf("│ ✅ Rimossa 1 canzone da %s\n", p.Name)
					} else {
						fmt.Printf("│ ✅ Rimosse %d canzoni da %s\n", len(tracksToRemove), p.Name)
					}

					// Mostra l'elenco delle canzoni rimosse
					fmt.Printf("│    🎵 Canzoni rimosse:\n")
					trackDetails, err := getTrackDetails(tracksToRemove)
					if err != nil {
						log.Warn("Errore nel recupero dettagli tracce rimosse", "error", err)
					} else {
						for _, track := range trackDetails {
							fmt.Printf("│       ↪ %s\n", track.Name)
						}
					}
				}
			} else {
				log.Info("Rimozione canzoni saltata per scelta utente", "playlistName", p.Name)
			}
		}

		// Separatore tra playlist
		fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────────")
		fmt.Println()
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║               ✅ AGGIORNAMENTO COMPLETATO ✅               ║")
	fmt.Println("║     Tutte le playlist collegate sono state aggiornate!     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	return nil
}
