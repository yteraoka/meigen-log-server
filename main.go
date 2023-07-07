package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const URL string = "https://meigen.doodlenote.net/api/json.php"

var version string

type Response struct {
	Meigen string `json:"meigen"`
	Author string `json:"auther"` // typo?
}

func getMeigen() (meigen, author string, err error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	r := make([]Response, 0)
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", "", err
	}
	return r[0].Meigen, r[0].Author, nil
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, version)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var err error

	nLogs := 1
	nLogsQuery, ok := r.URL.Query()["n"]

	if ok {
		nLogs, err = strconv.Atoi(nLogsQuery[0])
		if err != nil {
			nLogs = 1
			log.Warn().Str("uri", r.RequestURI).Str("method", r.Method).
				Err(err).Msgf("failed to atoi n parameter: %q", nLogsQuery[0])
		}
	}

	length := 0
	lengthQuery, ok := r.URL.Query()["length"]

	if ok {
		length, err = strconv.Atoi(lengthQuery[0])
		if err != nil {
			length = 0
			log.Warn().Str("uri", r.RequestURI).Str("method", r.Method).
				Err(err).Msgf("failed to atoi length parameter: %q", lengthQuery[0])
		}
	}

	for i := 0; i < nLogs; i++ {
		meigen := ""
		author := ""
		tmp := ""
		if length == 0 {
			meigen, author, err = getMeigen()
			if err != nil {
				log.Error().Err(err).Msg("failed to get meigen")
				break
			}
		} else if length > 0 {
			for {
				tmp, author, err = getMeigen()
				if err != nil {
					log.Error().Err(err).Msg("failed to get meigen")
					break
				}
				if len(meigen) == 0 {
					meigen = tmp + " by " + author
				} else {
					meigen = meigen + ";; " + tmp + " by " + author
				}
				if len(meigen) > length {
					break
				}
			}
			author = ""
		}

		u, _ := uuid.NewRandom()

		log.Info().Str("uri", r.RequestURI).Str("method", r.Method).
			Str("author", author).Str("uuid", u.String()).
			Str("length", strconv.Itoa(len(meigen))).Msg(meigen)
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.LevelFieldName = "severity"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	http.HandleFunc("/health", healthcheckHandler)
	http.HandleFunc("/", handler)

	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = "8080"
	}
	listenAddr := os.Getenv("LISTEN_ADDR")
	log.Debug().Msgf("listening %s:%s", listenAddr, listenPort)
	err := http.ListenAndServe(listenAddr+":"+listenPort, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
