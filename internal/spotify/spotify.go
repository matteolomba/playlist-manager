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

// Playlist struct used to store the playlist data in json files to backup and restore them
type Playlist struct {
	ID        api.ID   `json:"id"`
	Name      string   `json:"name"`
	TracksIDs []api.ID `json:"tracks"`
}

// IsAuthenticated returns true if the user is authenticated or false otherwise
func IsAuthenticated() bool {
	return authDone
}

// Init initializes the Spotify API. It needs to be called before any other function in this package
func Init() {
	err := getAuthVars()
	if err != nil {
		log.Fatal("Inizializzazione Spotify: recupero delle variabili per l'autenticazione", "error", err)
	}

	//Gin router setup
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

	//Server setup
	srv = &http.Server{
		Addr:    ":80",
		Handler: r,
	}

	// Auth endpoint
	r.GET("/api/auth", authEndpoint)

	//Check and create the required dir
	//Backup dir
	if _, err := os.Stat("data/backup"); os.IsNotExist(err) {
		os.MkdirAll("data/backup", 0644) // Create dir if not exists
		log.Info("Cartella data/backup creata")
	}
	//Auth dir
	if _, err := os.Stat("data/auth"); os.IsNotExist(err) {
		os.MkdirAll("data/auth", 0644) // Create dir if not exists
		log.Info("Cartella data/auth creata")
	}
	//Playlists dir
	if _, err := os.Stat("data/playlists"); os.IsNotExist(err) {
		os.MkdirAll("data/playlists", 0644) // Create dir if not exists
		log.Info("Cartella data/playlists creata")
	}
}

// -> Auth vars functions
// Struct for parsing the YAML file data/auth/auth.yaml, containing the auth vars
type authVarsModel struct {
	State   string `yaml:"csrfState"`
	AuthKey string `yaml:"authKey"`
}

/*
readAuthVars reads the auth vars (authKey for cookies and csrfState to avoid CSRF) from the data/auth/auth.yaml file
Returns the auth vars and an error, if present
*/
func readAuthVars() (vars *authVarsModel, err error) {
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
createAuthVars creates the auth vars (authKey for cookies and csrfState to avoid CSRF) and saves them to the data/auth/auth.yaml file
Returns the auth vars and an error, if present
*/
func createAuthVars() (vars *authVarsModel, err error) {
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

/*
getAuthVars reads the auth vars from the auth.yaml file and creates them if they don't exist
Returns an error, if present
*/
func getAuthVars() (err error) {
	//Reading auth vars,
	authVars, err = readAuthVars()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// If the file doesn't exist or is empty, create new auth vars
	if authVars.AuthKey == "" || authVars.State == "" {
		authVars, err = createAuthVars()
		if err != nil {
			return err
		}
	}
	return nil
}

//-> Auth and auth server functions

// startServer starts the gin server to receive the auth token and makes the program crash if there is an error
func startServer() {
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server per l'autenticazione chiuso con successo")
	} else if err != nil {
		log.Fatal("[Spotify] Chiusura server", "error", err)
	}
}

/*
Auth authenticates the user and saves the token in the data/auth/token.json file
Returns an error, if present
*/
func Auth() (err error) {
	authenticator = apiauth.New(apiauth.WithRedirectURL("http://localhost/api/auth"), apiauth.WithScopes(apiauth.ScopeUserReadPrivate, apiauth.ScopePlaylistReadPrivate, apiauth.ScopePlaylistReadCollaborative, apiauth.ScopePlaylistModifyPrivate))

	token, err := readAuthToken()
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

		go startServer()

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

// authEndpoint is the authentication endpoint for the gin (http) server
func authEndpoint(ctx *gin.Context) {
	token, err := authenticator.Token(context, authVars.State, ctx.Request)
	if err != nil {
		log.Error("Ottenimento del token da Spotify: ", "error", err)
		ctx.String(http.StatusNotFound, "Ottenimento del token da Spotify: "+err.Error()) //TODO: Farlo funzionare
		return
	}

	err = saveAuthToken(token)
	if err != nil {
		log.Error("Errore nel salvataggio del token per l'autenticazione: " + err.Error())
	}

	ctx.String(http.StatusOK, "Autenticazione completata con successo. Ora puoi chiudere questa pagina.")

	//Continue the auth process and send the client
	ch <- api.New(authenticator.Client(context, token))
}

/*
readAuthToken reads the auth token from the data/auth/token.json file.
Returns the token and an error, if present
*/
func readAuthToken() (token *oauth2.Token, err error) {
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

/*
saveAuthToken saves the auth token (given as parameter) to the data/auth/token.json file
Returns an error, if present
*/
func saveAuthToken(token *oauth2.Token) (err error) {
	data, err := json.Marshal(token)
	if err != nil {
		return
	}

	//Write file
	err = os.WriteFile("data/auth/token.json", data, 0644)
	if err != nil {
		return
	}
	log.Info("Token per l'autenticazione salvato in data/auth/token.json")
	return nil
}

//-> Playlist, track and user functions

// GetPlaylists returns the playlists of the authenticated user and an error, if present
func GetPlaylists() (pl []api.SimplePlaylist, err error) {
	pl = []api.SimplePlaylist{}
	user, err := client.CurrentUser(context)
	if err != nil {
		return
	}
	res, err := client.GetPlaylistsForUser(context, user.ID)
	return res.Playlists, err
}

// GetTracks returns the tracks of a playlist, given its ID, and an error, if present
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

// GetTrackIDs returns the IDs of the tracks (only music not podcasts) of a playlist, given its ID, and an error, if present
func GetTrackIDs(playlistID api.ID) (trackIDs []api.ID, err error) {
	tracklist := []api.PlaylistItem{}
	res, err := client.GetPlaylistItems(context, playlistID)
	if err != nil {
		return nil, err
	}
	tracklist = append(tracklist, res.Items...)
	for {
		err = client.NextPage(context, res)
		if err == api.ErrNoMorePages {
			break
		} else if err != nil {
			return nil, err
		}
		tracklist = append(tracklist, res.Items...)
	}
	//Get track IDs from the tracklist
	for _, t := range tracklist {
		if t.Track.Track.ID == "" {
			log.Warn("Brano non disponibile, potrebbe essere un podcast o un brano non disponibile su Spotify")
		} else {
			trackIDs = append(trackIDs, t.Track.Track.ID)
		}
	}
	return trackIDs, nil
}

// GetUserID returns the ID of the authenticated user and an error, if present
func GetUserID() (string, error) {
	user, err := client.CurrentUser(context)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

/*
AddTracksToPlaylist adds the tracks (given the ID) from trackList to playlist given its ID (playlistID)
Returns an error, if present
*/
func AddTracksToPlaylist(trackList []api.ID, playlistID api.ID) (err error) {
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
