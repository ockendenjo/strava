package strava

type Activity struct {
	ID        int64       `json:"id"`
	Name      string      `json:"name"`
	Map       PolylineMap `json:"map"`
	SportType string      `json:"sport_type"`
	GearID    string      `json:"gear_id"`
	StartDate string      `json:"start_date"`
}

type ActivityStream struct {
	LatLng LatLngData `json:"latlng"`
}

type LatLngData struct {
	Data [][]float64 `json:"data"`
}

type PolylineMap struct {
	ID       string `json:"id"`
	Polyline string `json:"polyline"`
}
