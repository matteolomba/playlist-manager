# playlist-manager-cli

Applicazione da terminale per gestire delle playlist su [Spotify](https://spotify.com)

⌨️ Codice scritto in inglese, prima o poi metterò la documentazione e le stringhe del programma anche in inglese
<br> 🇬🇧 The code is already written in English, someday I will put the documentation and program strings in English as well

⚠️ Quello che c'è dovrebbe funzionare ma non è garantito, ho effettuato un test ridotto

🐛 Se trovi un problema o vorresti una nuova funzionalità apri un [issue](https://github.com/matteolomba/playlist-manager-cli/issues) o una [pull request](https://github.com/matteolomba/playlist-manager-cli/pulls)

## Cosa ci puoi fare

- Il backup e ripristino di playlist da Spotify come file JSON

- Gestire delle playlist collegate, cos'è una playlist collegata?
<br> Una playlist collegata è una playlist che contiene tutte le canzoni di almeno 2 playlist, con la conseguente aggiunta/rimozione (dalla playlist di destinazione) delle canzoni che sono state aggiunte/rimosse dalle playlist originali. Per effettuare l'aggiornamento bisogna usare la scelta dedicata nel menu

## Primo avvio e configurazione

Per utilizzare l'applicazione è necessario creare un'applicazione su Spotify e ottenere le credenziali per l'accesso all'API, ottienile [qui](https://developer.spotify.com/dashboard)

Una volta ottenute, vanno inserite nel file `.env` nella root del progetto/eseguibile. L'esempio e le informazioni sono nel file `.env.example`, basta rinominarlo in `.env` e inserire i dati

## Crediti

- [zmb3's spotify wrapper](https://github.com/zmb3/spotify/) - Go wrapper used to interact with [Spotify's Web API](https://developer.spotify.com/documentation/web-api)
