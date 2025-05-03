package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/hwanbin/wanpm/internal/data"
)

func (app *application) forwardGeocodeHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	accessToken := os.Getenv("MAPBOX_GEOCODE_TOKEN")
	proximity := "ip"
	searchText := app.readString(qs, "q", "")

	t := &url.URL{
		Path: searchText,
	}

	geocodeURL := "https://api.mapbox.com/search/geocode/v6/forward"
	requestURL := fmt.Sprintf("%s?q=%s&proximity=%s&access_token=%s", geocodeURL, t.String(), proximity, accessToken)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		// add to error bag
		fmt.Printf("client: could not create request: %s\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return
	}

	var input data.FeatureCollection

	err = app.readMapboxJSON(res, &input)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"result": input}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
