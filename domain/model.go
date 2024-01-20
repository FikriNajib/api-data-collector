package domain

import "time"

type DataCollectorRequest struct {
	UserID           interface{} `json:"user_id"`
	Service          string      `json:"service"`
	ContentType      string      `json:"content_type"`
	ContentID        int         `json:"content_id"`
	Action           string      `json:"action_user"`
	Title            string      `json:"title"`
	Name             string      `json:"name"`
	DeviceID         string      `json:"device_id"`
	Platform         string      `json:"platform"`
	DateTime         time.Time   `json:"date_time"`
	Hashtag          []string    `json:"hashtag"`
	Pillar           string      `json:"pillar"`
	Duration         float64     `json:"duration"`
	IpAddress        string      `json:"ip_address"`
	CreatorID        int         `json:"creator_id"`
	VideoDuration    int         `json:"video_duration"`
	IsIncognito      bool        `json:"is_incognito"`
	IsEmbeddedIframe bool        `json:"is_iframe"`
	UserAgent        string      `json:"user_agent"`
}
