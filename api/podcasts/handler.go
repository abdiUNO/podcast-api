package podcasts

import (
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go-podcast-api/utils/response"
	"log"
	"net/http"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

var GetAllPodcasts = func(w http.ResponseWriter, r *http.Request) {

	//token := r.Context().Value("token").(*auth.Token)
	//user := auth.GetUser(token.UserId)

	//ctx := context.Background()
	//
	//cachedPodcasts, err := redisClient.Get(ctx, "products").Bytes()

	podcasts, err := AllPodcasts()

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"podcasts": podcasts,
	})
}

var FindPodcasts = func(w http.ResponseWriter, r *http.Request) {
	//token := r.Context().Value("token").(*Token)
	query := r.FormValue("q")

	podcasts, err := SearchPodcastByName(query)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"podcasts": podcasts,
	})

}

var GetPodcast = func(w http.ResponseWriter, r *http.Request) {
	//token := r.Context().Value("token").(*Token)
	params := mux.Vars(r)
	podId := params["id"]
	log.Println(podId)
	podcast, feed, err := FindPodcastById(podId)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"podcast": podcast,
		"feed":    feed,
	})

}

var GetTopPodcasts = func(w http.ResponseWriter, r *http.Request) {
	//token := r.Context().Value("token").(*Token)
	podcasts, err := TopPodcasts()
	if err != nil {
		log.Println(err)
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"podcasts": podcasts,
	})

}
