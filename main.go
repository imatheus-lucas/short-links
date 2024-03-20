package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/imatheus-lucas/short-links/database"
	"github.com/imatheus-lucas/short-links/redisClient"
	"github.com/redis/go-redis/v9"
)

type Body struct {
	LongUrl  string `json:"long_url"`
	ShortUrl string `json:"short_url"`
}

type Reply = map[string]string
type Link struct {
	LinkId any     `json:"linkId"`
	Score  float64 `json:"score"`
}

func main() {

	db := database.Init()

	database.Migrate(db)
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /links", func(w http.ResponseWriter, r *http.Request) {

		var body Body
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			response(w, Reply{"error": err.Error()}, http.StatusBadRequest)
			return
		}
		var id int
		err := db.QueryRow("INSERT INTO tb_links (long_url, short_url) VALUES ($1, $2) RETURNING id", body.LongUrl, body.ShortUrl).Scan(&id)

		if err != nil {
			response(w, Reply{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		response(w, Reply{"id": fmt.Sprintf("%d", id)}, http.StatusCreated)
	})
	mux.HandleFunc("GET /{link}", func(w http.ResponseWriter, r *http.Request) {
		link := r.PathValue("link")

		if link == "" {
			response(w, Reply{"error": "No link provided"}, http.StatusBadRequest)
			return
		}
		var longUrl string
		var id int
		err := db.QueryRow("SELECT long_url, id FROM tb_links WHERE short_url = $1", link).Scan(&longUrl, &id)

		if err != nil {
			response(w, Reply{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		if longUrl == "" {
			response(w, Reply{"error": "Link not found"}, http.StatusNotFound)
			return
		}
		ctx := context.Background()
		rdb := redisClient.Init()

		rdb.ZIncrBy(ctx, "metrics", 1, strconv.Itoa(id))

		http.Redirect(w, r, longUrl, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("GET /metrics/{link_id}", func(w http.ResponseWriter, r *http.Request) {
		link := r.PathValue("link_id")

		if link == "" {
			response(w, Reply{"error": "No link provided"}, http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		rdb := redisClient.Init()

		result, err := rdb.ZRangeByScoreWithScores(ctx, "metrics", &redis.ZRangeBy{
			Min: "0",
			Max: "+inf",
		}).Result()

		if err != nil {
			response(w, Reply{"error": err.Error()}, http.StatusBadRequest)
			return
		}
		//convert result

		var links []Link

		for _, link := range result {

			links = append(links, Link{
				LinkId: link.Member,
				Score:  link.Score,
			})
		}

		response(w, links, http.StatusOK)

	})

	http.ListenAndServe(":3333", mux)
}

func response(w http.ResponseWriter, reply any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(reply)
}
