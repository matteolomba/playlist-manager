package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"playlist-manager/internal/spotify"
	"playlist-manager/pkg/utils"
	"time"

	"github.com/savioxavier/termlink"
	api "github.com/zmb3/spotify/v2"

	log "playlist-manager/pkg/logger"
)

var userID string

const VERSION = "0.3.6"

func Display() (err error) {
	log.Info("Avvio di Playlist Manager", "version", VERSION)
	options := []string{
		"Visualizza le playlist del tuo account",
		"Visualizza i brani di una playlist del tuo account",
		"Salva una playlist (Backup)",
		"Salva tutte le playlist (Backup)",
		"Carica una playlist (Restore)",
		"Visualizza e gestisci le playlist collegate",
	}

	err = spotify.Auth()
	if err != nil {
		log.Error("Errore durante l'autenticazione Spotify", "error", err)
		return err
	}
	log.Info("Autenticazione Spotify completata con successo")

	//Clear terminal after auth
	utils.ClearTerminal()

	for {
		fmt.Println("========================================================")
		fmt.Println("🎧 Playlist Manager " + VERSION + " 🎧\n✨ Sviluppato da " + termlink.ColorLink("Matteo Lombardi", "https://github.com/matteolomba", "italic yellow") + " ✨")
		fmt.Println("========================================================")
		displayAuthStatus()
		fmt.Println("========================================================")
		fmt.Println("🏠 -> Menù Principale <- 🏠")
		fmt.Println("========================================================")
		fmt.Printf("🚪 0. Esci\n")
		optionEmojis := []string{"📋", "🎵", "💾", "📁", "🔄", "🔗"}
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
			log.Info("L'utente ha scelto di uscire dall'applicazione", "userID", userID)
			fmt.Println("👋 Ciao! Esco dall'applicazione...")
			return nil

		case 1: // Get playlists
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto la visualizzazione delle playlist", "userID", userID)
			pl, err := spotify.GetPlaylists()
			if err != nil {
				log.Error("Errore nel recupero delle playlist", "error", err, "userID", userID)
				return err
			}
			log.Info("Playlist recuperate con successo", "count", len(pl), "userID", userID)
			fmt.Println("\n🎼 Le tue playlist:")
			fmt.Println("==================")
			for i, p := range pl {
				fmt.Printf("🎵 %d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Printf("\n⏎ Premi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 2: // Get tracks from a playlist
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto la visualizzazione dei brani di una playlist", "userID", userID)
			pl, err := spotify.GetPlaylists()
			if err != nil {
				log.Error("Errore nel recupero delle playlist per la visualizzazione dei brani", "error", err, "userID", userID)
				return err
			}
			fmt.Println("\n🎼 Seleziona una playlist:")
			fmt.Println("=========================")
			for i, p := range pl {
				fmt.Printf("🎵 %d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Println("🔙 0. Annulla")
			fmt.Print("\n🔍 Inserisci il numero della playlist di cui vuoi visualizzare i brani: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				log.Error("Errore nella lettura della selezione playlist per brani", "error", err, "userID", userID)
				return err
			}
			if sel == 0 {
				log.Info("L'utente ha annullato la visualizzazione dei brani", "userID", userID)
				break
			}
			if sel < 1 || sel > len(pl) {
				log.Warn("Selezione non valida per visualizzare i brani", "selection", sel, "max", len(pl), "userID", userID)
				fmt.Println("❌ Selezione non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}
			selectedPlaylist := pl[sel-1]
			utils.ClearTerminal()
			log.Info("L'utente ha selezionato una playlist per visualizzazione brani", "playlistName", selectedPlaylist.Name, "playlistID", selectedPlaylist.ID, "userID", userID)
			tracks, err := spotify.GetTracks(selectedPlaylist.ID)
			if err != nil {
				log.Error("Errore nel recupero dei brani della playlist", "error", err, "playlistName", selectedPlaylist.Name, "playlistID", selectedPlaylist.ID, "userID", userID)
				return err
			}
			log.Info("Brani della playlist recuperati con successo", "trackCount", len(tracks), "playlistName", selectedPlaylist.Name, "userID", userID)
			fmt.Printf("\n🎵 Brani della playlist '%s':\n", selectedPlaylist.Name)
			fmt.Println("=======================================")
			for i, t := range tracks {
				if t.Track.Track.ID == "" {
					fmt.Printf("⚠️ %d. Brano non disponibile (potrebbe essere un podcast)\n", i+1)
					log.Warn("Brano non disponibile, potrebbe essere un podcast o un brano non disponibile su Spotify")
				} else {
					text := fmt.Sprintf("🎶 %d. %s di ", i+1, t.Track.Track.Name)
					for _, a := range t.Track.Track.Artists {
						text += a.Name + ", "
					}
					text = text[:len(text)-2]
					fmt.Println(text)
				}
			}
			fmt.Printf("\n⏎ Premi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 3: // Save playlist (Backup) to JSON file
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto il backup di una singola playlist", "userID", userID)
			//Get playlists
			pl, err := spotify.GetPlaylists()
			if err != nil {
				log.Error("Errore nel recupero delle playlist per backup singolo", "error", err, "userID", userID)
				return err
			}
			fmt.Println("\n💾 Seleziona la playlist da salvare:")
			fmt.Println("====================================")
			for i, p := range pl {
				fmt.Printf("🎵 %d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Println("🔙 0. Annulla")

			//Select playlist
			fmt.Print("\n💾 Inserisci il numero della playlist di cui vuoi fare il backup: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				log.Error("Errore nella lettura della playlist selezionata per backup", "error", err, "userID", userID)
				return err
			}
			if sel == 0 {
				log.Info("L'utente ha annullato il backup della playlist", "userID", userID)
				break
			}
			if sel < 1 || sel > len(pl) {
				log.Warn("Selezione non valida per il backup della playlist", "selection", sel, "max", len(pl), "userID", userID)
				fmt.Println("❌ Selezione non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			selectedPlaylist := pl[sel-1]
			utils.ClearTerminal()
			log.Info("Inizio backup playlist singola", "playlistName", selectedPlaylist.Name, "playlistID", selectedPlaylist.ID, "userID", userID)
			//Save playlist
			backupDir, err := savePlaylistAsJSON(selectedPlaylist, userID)
			if err != nil {
				log.Error("Errore durante il backup della playlist", "error", err, "playlistName", selectedPlaylist.Name, "playlistID", selectedPlaylist.ID, "userID", userID, "backupDir", backupDir)
				return err
			}
			log.Info("Backup playlist completato con successo", "playlistName", selectedPlaylist.Name, "playlistID", selectedPlaylist.ID, "userID", userID, "backupDir", backupDir)
			fmt.Println("✅ Playlist salvata in:", backupDir)
			fmt.Printf("\n⏎ Premi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 4: // Backup all personal playlists to JSON files
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto il backup di tutte le playlist personali", "userID", userID)
			//Get playlists
			pl, err := spotify.GetPlaylists()
			if err != nil {
				log.Error("Errore nel recupero delle playlist per backup multiplo", "error", err, "userID", userID)
				return err
			}

			personalPlaylistsCount := 0
			for _, p := range pl {
				if p.Owner.ID == userID {
					personalPlaylistsCount++
				}
			}
			log.Info("Playlist personali da salvare ottenute", "count", personalPlaylistsCount, "userID", userID)

			today := time.Now().Format("2006-01-02")
			fmt.Printf("💾 Verranno salvate %d playlist personali in:\n📂 %s\n\n", personalPlaylistsCount, "data/backup/"+userID+"/"+today+"/")
			fmt.Println("⏳ Avvio backup...")
			savedCount := 0
			for _, p := range pl {
				//Process only personal playlists
				if p.Owner.ID == userID {
					//Save playlist
					log.Info("Inizio backup playlist", "playlistName", p.Name, "playlistID", p.ID, "userID", userID)
					_, err = savePlaylistAsJSON(p, userID)
					if err != nil {
						log.Error("Errore durante il backup della playlist", "error", err, "playlistName", p.Name, "playlistID", p.ID, "userID", userID)
						return err
					}
					savedCount++
					log.Info("Backup playlist completato", "playlistName", p.Name, "playlistID", p.ID, "userID", userID)
				}
			}
			log.Info("Backup multiplo completato", "totalSaved", savedCount, "userID", userID)
			fmt.Printf("✅ Backup completato! Salvate %d playlist personali.\n\n⏎ Premi invio per tornare al menu...", savedCount)
			fmt.Scanf("\n\n")

		case 5: // Restore playlist from JSON file
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto il ripristino di una playlist", "userID", userID)

			// Prima scelta: playlist proprie o di altri
			fmt.Println("\n🔄 Che playlist vuoi ripristinare?")
			fmt.Println("=====================================")
			fmt.Println("👤 1. Una delle mie personali")
			fmt.Println("👥 2. Una di un altro utente")
			fmt.Println("🔙 0. Torna al menu")
			fmt.Print("\n❓ Scelta: ")

			var ownerChoice int
			_, err := fmt.Scan(&ownerChoice)
			if err != nil {
				log.Error("Errore nella lettura della scelta sul proprietario", "error", err, "userID", userID)
				return err
			}

			if ownerChoice == 0 {
				log.Info("L'utente ha annullato il ripristino", "userID", userID)
				break
			}

			var backupDir string
			var isPersonal bool

			switch ownerChoice {
			case 1:
				// Playlist personali
				utils.ClearTerminal()
				backupDir = "data/backup/" + userID
				isPersonal = true
				log.Info("L'utente ha scelto di ripristinare una playlist personale", "userID", userID)
			case 2:
				// Playlist di altri (nella cartella altre dell'utente)
				utils.ClearTerminal()
				backupDir = "data/backup/" + userID + "/altre"
				isPersonal = false
				log.Info("L'utente ha scelto di ripristinare una playlist di un altro utente", "userID", userID, "backupDir", backupDir)
			default:
				log.Warn("Scelta del proprietario non valida", "choice", ownerChoice, "userID", userID)
				fmt.Println("❌ Scelta non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				continue // Torna al menu principale invece di break
			}

			// Verifica esistenza cartella
			if _, err := os.Stat(backupDir); os.IsNotExist(err) {
				log.Warn("Nessun backup trovato", "userID", userID, "backupDir", backupDir, "isPersonal", isPersonal)
				if isPersonal {
					fmt.Println("❌ Nessun backup di tue playlist trovato")
				} else {
					fmt.Println("❌ Nessun backup di playlist di altri trovato")
				}
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Get date directories
			dateDirs, err := os.ReadDir(backupDir)
			if err != nil {
				log.Error("Errore nella lettura delle cartelle di backup", "error", err, "userID", userID, "backupDir", backupDir)
				return err
			}

			if len(dateDirs) == 0 {
				log.Info("Nessuna cartella di backup trovata", "userID", userID, "backupDir", backupDir)
				if isPersonal {
					fmt.Println("Nessun backup trovato per l'utente corrente")
				} else {
					fmt.Println("Nessun backup di playlist di altri trovato")
				}
				fmt.Printf("\nPremi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			log.Info("Cartelle di backup trovate", "count", len(dateDirs), "userID", userID)
			fmt.Println("\n📅 Seleziona da quale data ripristinare:")
			fmt.Println("=====================================")
			validDateDirs := []os.DirEntry{}
			dateIndex := 1
			for _, d := range dateDirs {
				if d.IsDir() {
					// Se siamo nelle playlist personali, nascondi la cartella "altre"
					if isPersonal && d.Name() == "altre" {
						continue
					}
					fmt.Printf("📆 %d. %s\n", dateIndex, d.Name())
					validDateDirs = append(validDateDirs, d)
					dateIndex++
				}
			}
			fmt.Println("🔙 0. Torna al menu")

			if len(validDateDirs) == 0 {
				log.Warn("Nessuna cartella di backup valida trovata", "userID", userID)
				fmt.Println("❌ Nessuna cartella di backup valida trovata.")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Select date
			fmt.Print("\n📅 Inserisci il numero della data: ")
			var dateSelect int
			_, err = fmt.Scan(&dateSelect)
			if err != nil {
				log.Error("Errore nella lettura della selezione data", "error", err, "userID", userID)
				return err
			}
			if dateSelect == 0 {
				log.Info("L'utente ha annullato la selezione della data", "userID", userID)
				break
			}
			if dateSelect < 1 || dateSelect > len(validDateDirs) {
				log.Warn("Selezione data non valida per il ripristino", "selection", dateSelect, "max", len(validDateDirs), "userID", userID)
				fmt.Println("❌ Selezione non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			selectedDateDir := validDateDirs[dateSelect-1].Name()
			utils.ClearTerminal()
			playlistDir := backupDir + "/" + selectedDateDir
			log.Info("Data selezionata per il ripristino", "date", selectedDateDir, "userID", userID, "isPersonal", isPersonal)

			//Get playlist files from selected date
			files, err := os.ReadDir(playlistDir)
			if err != nil {
				return err
			}

			if len(files) == 0 {
				fmt.Printf("❌ Nessun backup trovato per la data %s.\n", selectedDateDir)
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			fmt.Printf("\n🎵 Playlist salvate il %s:\n", selectedDateDir)
			fmt.Println("=======================================")
			validPlaylistFiles := []os.DirEntry{}
			playlistIndex := 1
			for _, f := range files {
				if !f.IsDir() && len(f.Name()) > 5 && f.Name()[len(f.Name())-5:] == ".json" {
					//Read file to get playlist name
					tempData, err := os.ReadFile(playlistDir + "/" + f.Name())
					if err != nil {
						log.Warn("Impossibile leggere il file: " + f.Name())
						continue
					}

					//Parse JSON
					var tempPl spotify.Playlist
					err = json.Unmarshal(tempData, &tempPl)
					if err != nil {
						log.Warn("Impossibile parsare il file JSON: " + f.Name())
						continue
					}

					fmt.Printf("🎵 %d. %s (%s)\n", playlistIndex, tempPl.Name, f.Name())
					validPlaylistFiles = append(validPlaylistFiles, f)
					playlistIndex++
				}
			}

			fmt.Println("🔙 0. Torna al menu")
			if len(validPlaylistFiles) == 0 {
				fmt.Printf("❌ Nessun file di backup valido trovato per la data %s.\n", selectedDateDir)
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Select playlist file
			fmt.Print("\n🔄 Inserisci il numero della playlist da caricare: ")
			var playlistSelect int
			_, err = fmt.Scan(&playlistSelect)
			if err != nil {
				return err
			}
			if playlistSelect == 0 {
				break
			}
			if playlistSelect < 1 || playlistSelect > len(validPlaylistFiles) {
				fmt.Println("❌ Selezione non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Read selected playlist file
			selectedPlaylistFile := validPlaylistFiles[playlistSelect-1].Name()
			utils.ClearTerminal()
			data, err := os.ReadFile(playlistDir + "/" + selectedPlaylistFile)
			if err != nil {
				return err
			}

			//Parse JSON
			var playlist spotify.Playlist
			err = json.Unmarshal(data, &playlist)
			if err != nil {
				return err
			}

			//Get current playlists to restore into
			pl, err := spotify.GetPlaylists()
			if err != nil {
				return err
			}

			fmt.Printf("\n🎯 Seleziona la playlist di destinazione per '%s':\n", playlist.Name)
			fmt.Println("======================================================")
			for i, p := range pl {
				fmt.Printf("🎵 %d. %s - %s\n", i+1, p.Name, p.ID)
			}
			fmt.Println("🔙 0. Annulla")

			fmt.Print("\n📥 Inserisci il numero della playlist in cui caricare i brani: ")
			var sel int
			_, err = fmt.Scan(&sel)
			if err != nil {
				return err
			}
			if sel == 0 {
				log.Info("L'utente ha annullato la selezione della playlist di destinazione", "userID", userID)
				break
			}
			if sel < 1 || sel > len(pl) {
				fmt.Println("❌ Selezione non valida")
				fmt.Printf("\n⏎ Premi invio per tornare al menu...")
				fmt.Scanf("\n\n")
				break
			}

			//Restore playlist
			utils.ClearTerminal()
			fmt.Printf("⏳ Ripristino di '%s' in corso...\n", playlist.Name)
			err = spotify.AddTracksToPlaylist(playlist.TrackIDs, pl[sel-1].ID)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Playlist '%s' caricata con successo in '%s'!\n", playlist.Name, pl[sel-1].Name)
			fmt.Printf("\n⏎ Premi invio per tornare al menu...")
			fmt.Scanf("\n\n")

		case 6: // Manage linked playlists
			utils.ClearTerminal()
			log.Info("L'utente ha richiesto la gestione delle playlist collegate", "userID", userID)
			err := linkedMenu()
			if err != nil {
				log.Error("Errore nella gestione delle playlist collegate", "error", err, "userID", userID)
				return err
			}

		default:
			log.Warn("Scelta menu non valida", "selection", sel, "userID", userID)
			fmt.Println("❌ Scelta non valida o non ancora implementata")
		}
		utils.ClearTerminal()
	}
}

func displayAuthStatus() {
	fmt.Print("🔐 Autenticato: ")
	if spotify.IsAuthenticated() {
		fmt.Println("✅")

		//Display user ID
		if userID == "" {
			var err error
			userID, err = spotify.GetUserID()
			if err != nil {
				log.Error("Errore nel recupero dell'ID dell'utente", "error", err)
			} else {
				log.Info("UserID recuperato con successo", "userID", userID)
				fmt.Println("👤 Utente:", userID)
			}
		} else {
			fmt.Println("👤 Utente:", userID)
		}

	} else {
		log.Warn("Utente non autenticato")
		fmt.Println("❌")
	}
}

// savePlaylistAsJSON salva una playlist mostrando un'animazione di caricamento
func savePlaylistAsJSON(playlist api.SimplePlaylist, userID string) (backupDir string, err error) {
	log.Info("Inizio salvataggio playlist (con animazione)", "playlistName", playlist.Name, "playlistID", playlist.ID, "userID", userID)
	fmt.Printf("💾 Salvataggio di '%s' in corso ", playlist.Name)

	// Esegui il salvataggio in una goroutine
	done := make(chan bool)
	errChan := make(chan error)

	go func() {
		backupDir, err = spotify.SavePlaylistAsJSON(playlist, userID)
		if err != nil {
			log.Error("Errore durante il salvataggio", "error", err, "playlistName", playlist.Name, "playlistID", playlist.ID, "userID", userID, "backupDir", backupDir)
			errChan <- err
		} else {
			log.Info("Salvataggio completato con successo", "playlistName", playlist.Name, "playlistID", playlist.ID, "userID", userID, "backupDir", backupDir)
			done <- true
		}
	}()

	// Animazione a cerchio
	spinChars := []string{"|", "/", "-", "\\"}
	i := 0

	for {
		select {
		case <-done:
			fmt.Printf("\n✅ Playlist '%s' salvata come %s.json\n\n", playlist.Name, string(playlist.ID))
			return backupDir, nil
		case err := <-errChan:
			fmt.Printf("\n❌ Errore nel salvataggio della playlist '%s': %v\n\n", playlist.Name, err)
			return "", err
		default:
			fmt.Printf("%s", spinChars[i%4])
			i++
			time.Sleep(200 * time.Millisecond)
			fmt.Print("\b")
		}
	}
}
