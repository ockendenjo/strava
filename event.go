package strava

type WebhookEvent struct {
	ObjectType     string          `json:"object_type"`
	ObjectID       int64           `json:"object_id"`
	AspectType     string          `json:"aspect_type"`
	Updates        *WebhookUpdates `json:"updates"`
	OwnerID        int64           `json:"owner_id"`
	SubscriptionID int             `json:"subscription_id"`
	EventTime      int64           `json:"event_time"`
}

type WebhookUpdates struct {
	Title      *string `json:"title"`
	Type       *string `json:"type"`
	Private    *string `json:"private"`
	Authorized *string `json:"authorized"`
}
