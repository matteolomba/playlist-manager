package spotify

import (
	ctx "context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"playlist-manager/pkg/utils"
	"time"

	log "playlist-manager/pkg/logger"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	api "github.com/zmb3/spotify/v2"
	apiauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

var (
	authVars      *authVarsModel
	authDone      bool
	authenticator *apiauth.Authenticator
	client        *api.Client
	context       ctx.Context  = ctx.Background()
	srv           *http.Server // Gin server used to receive the auth token
	ch            = make(chan *api.Client)
)

type Playlist struct {
	ID        api.ID   `json:"id"`
	Name      string   `json:"name"`
	TracksIDs []api.ID `json:"tracks"`
}

func IsAuthenticated() bool {
	return authDone
}

// Deve essere chiamato prima di qualsiasi altra funzione di questo package
func Init() {
	err := recuperaAuthVars()
	if err != nil {
		log.Fatal("Inizializzazione Spotify: recupero delle variabili per l'autenticazione", "error", err)
	}
	//Definizione del router di Gin e delle impostazioni correlate
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	store := cookie.NewStore([]byte(authVars.AuthKey))
	r.Use(sessions.Sessions("session", store))
	r.Use(csrf.Middleware(csrf.Options{
		Secret: authVars.State,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	//Definizione del server
	srv = &http.Server{
		Addr:    ":80",
		Handler: r,
	}

	// Endpoint per autenticazione
	r.GET("/api/auth", authEndpoint)
}

// -> Parte per la gestione delle variabili per l'autenticazione
// Struct per parsing del file YAML data/auth/auth.yaml, contenente le variabili per l'autenticazione
type authVarsModel struct {
	State   string `yaml:"csrfState"`
	AuthKey string `yaml:"authKey"`
}

/*
Legge le variabili per l'autenticazione (authKey per i cookie e csrfState per evitare CSRF) dal file data/auth/auth.yaml
Ritorna le variabili per l'autenticazione e un eventuale errore
*/
func leggiAuthVars() (vars *authVarsModel, err error) {
	vars = &authVarsModel{}
	data, err := os.ReadFile("data/auth/auth.yaml")
	if errors.Is(err, os.ErrNotExist) {
		log.Warn("File data/auth/auth.yaml non esistente. Verranno generate nuove variabili per l'autenticazione.")
		return
	} else if err != nil {
		return
	}

	err = yaml.Unmarshal(data, vars)
	if err != nil {
		return
	}
	log.Info("Variabili per l'autenticazione lette da data/auth/auth.yaml")
	return vars, nil
}

/*
Crea le variabili per l'autenticazione (authKey per i cookie e csrfState per evitare CSRF) e le salva sul file data/auth/auth.yaml
Ritorna le variabili per l'autenticazione e un eventuale errore
*/
func creaAuthVars() (vars *authVarsModel, err error) {
	csrfState := utils.RandomString(16)
	authKey := utils.RandomString(16)
	vars = &authVarsModel{
		State:   csrfState,
		AuthKey: authKey,
	}
	data, err := yaml.Marshal(vars)
	if err != nil {
		return
	}

	err = os.WriteFile("data/auth/auth.yaml", data, 0644)
	if err != nil {
		return
	}
	log.Info("Variabili per l'autenticazione generate e salvate in data/auth/auth.yaml")
	return vars, nil
}

func recuperaAuthVars() (err error) {
	//Lettura delle variabili per l'autenticazione
	authVars, err = leggiAuthVars()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if authVars.AuthKey == "" || authVars.State == "" {
		authVars, err = creaAuthVars()
		if err != nil {
			return err
		}
	}
	return nil
}

//-> Auth and auth server functions

// Funzione per avviare il server Gin e gestire un eventuale errore
func avviaServer() {
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server per l'autenticazione chiuso con successo")
	} else if err != nil {
		log.Fatal("[Spotify] Chiusura server", "error", err)
	}
}

// Auth: Funzione per l'autenticazione con Spotify
func Auth() (err error) {
	authenticator = apiauth.New(apiauth.WithRedirectURL("http://localhost/api/auth"), apiauth.WithScopes(apiauth.ScopeUserReadPrivate, apiauth.ScopePlaylistReadPrivate, apiauth.ScopePlaylistReadCollaborative, apiauth.ScopePlaylistModifyPrivate))

	token, err := leggiAuthToken()
	if err == nil {
		if token.Valid() {
			//Token expired
			log.Info("Token per l'autenticazione scaduto. Verrà riefettuata l'autenticazione.")
			token, err = authenticator.RefreshToken(context, token)
			if err != nil {
				log.Error("Errore nel refresh del token per l'autenticazione: " + err.Error())
				return err
			}
		}
		client = api.New(authenticator.Client(context, token))
		authDone = true
		return nil

	} else {
		//authURL generation
		authURL := authenticator.AuthURL(authVars.State)
		log.Info("URL per l'autenticazione generata")

		fmt.Println("Apri questo URL per autenticarti: " + authURL)
		if authURL == "" {
			fmt.Println("Errore: URL per l'autenticazione non disponibile")
			return errors.New("URL di autenticazione non disponibile")
		}

		// Starting the server
		log.Info("Avvio il server per l'autenticazione...")

		go avviaServer()

		// Wait for auth to complete
		client = <-ch

		// Auth completed
		log.Info("Autenticazione completata")

		// Closing the server
		shutdownCtx, shutdownRelease := ctx.WithTimeout(ctx.Background(), 10*time.Second)
		defer shutdownRelease()

		err = srv.Shutdown(shutdownCtx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Errore nella chiusura del server: " + err.Error())
		}

		authDone = true
		return nil
	}
}

// Authentication endpoint for the gin(http) server
func authEndpoint(ctx *gin.Context) {
	token, err := authenticator.Token(context, authVars.State, ctx.Request)
	if err != nil {
		log.Error("Ottenimento del token da Spotify: ", "error", err)
		ctx.String(http.StatusNotFound, "Ottenimento del token da Spotify: "+err.Error()) //TODO: Farlo funzionare
		return
	}

	err = salvaAuthToken(token)
	if err != nil {
		log.Error("Errore nel salvataggio del token per l'autenticazione: " + err.Error())
	}

	ctx.String(http.StatusOK, "Autenticazione completata con successo. Ora puoi chiudere questa pagina.")

	//Continue the auth process and send the client
	ch <- api.New(authenticator.Client(context, token))
}

//-> Parte per la gestione del token per l'autenticazione

func leggiAuthToken() (token *oauth2.Token, err error) {
	token = &oauth2.Token{}
	data, err := os.ReadFile("data/auth/token.json")
	if errors.Is(err, os.ErrNotExist) {
		log.Warn("File data/auth/token.json non esistente. Verrà riefettuata l'autenticazione.")
		return
	} else if err != nil {
		return
	}

	err = json.Unmarshal(data, token)
	if err != nil {
		return
	}
	log.Info("Token per l'autenticazione letto da data/auth/token.json")
	return token, nil
}

func salvaAuthToken(token *oauth2.Token) (err error) {
	data, err := json.Marshal(token)
	if err != nil {
		return
	}

	//Dir exists?
	if _, err := os.Stat("data/auth"); os.IsNotExist(err) {
		os.MkdirAll("data/auth", 0644) // Create dir if not exists
	}

	//Write file
	err = os.WriteFile("data/auth/token.json", data, 0644)
	if err != nil {
		return
	}
	log.Info("Token per l'autenticazione salvato in data/auth/token.json")
	return nil
}

/*
Funzione per ottenere le playlist dell'utente, è necessario che l'utente sia autenticato (con funzione Auth())
Ritorna la lista delle playlist e un eventuale errore
*/
func GetPlaylists() (pl []api.SimplePlaylist, err error) {
	pl = []api.SimplePlaylist{}
	user, err := client.CurrentUser(context)
	if err != nil {
		return
	}
	res, err := client.GetPlaylistsForUser(context, user.ID)
	return res.Playlists, err
}

func GetTracks(playlistID api.ID) ([]api.PlaylistItem, error) {
	tracklist := []api.PlaylistItem{}
	res, err := client.GetPlaylistItems(context, playlistID)
	if err != nil {
		return nil, err
	}
	tracklist = append(tracklist, res.Items...)
	for {
		err = client.NextPage(context, res)
		if err == api.ErrNoMorePages {
			return tracklist, nil
		} else if err != nil {
			return nil, err
		}
		tracklist = append(tracklist, res.Items...)
	}
}

func GetUserID() (string, error) {
	user, err := client.CurrentUser(context)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func RestorePlaylist(trackList []api.ID, playlistID api.ID) (err error) {
	if len(trackList) > 100 {
		for i := 0; i < len(trackList); i += 100 {
			var tempTrackList []api.ID
			if i+100 > len(trackList) {
				tempTrackList = trackList[i:]
			} else {
				tempTrackList = trackList[i : i+100]
			}
			_, err = client.AddTracksToPlaylist(context, playlistID, tempTrackList...)
			if err != nil {
				return err
			}
		}
	} else {
		_, err = client.AddTracksToPlaylist(context, playlistID, trackList...)
		if err != nil {
			return err
		}
	}
	return nil
}
