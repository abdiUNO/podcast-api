package podcasts

import (
	"github.com/gorilla/mux"
	"go-podcast-api/utils/response"
	"log"
	"net/http"
)

var GetAllPodcasts = func(w http.ResponseWriter, r *http.Request) {

	//token := r.Context().Value("token").(*auth.Token)
	//user := auth.GetUser(token.UserId)

	podcasts, err := FindAllPodcasts()

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

	podcasts, err := QueryPodcast(query)
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
	podcast, feed, err := FindPodcast(podId)
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
