package discordrpcgenerator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hackirby/discordgo"
)

var (
	discordIdRegex = regexp.MustCompile(`^\d{17,19}$`)
	spotifyIdRegex = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)
)

// presence holds data for a Discord presence update (opcode 3)
type presence struct {
	data discordgo.UpdateStatusData
}

// NewPresence creates a new presence instance
func NewPresence() *presence {
	return &presence{
		data: discordgo.UpdateStatusData{
			AFK:       false,
			IdleSince: 0,
		},
	}
}

// Data returns the presence data
func (p *presence) Data() discordgo.UpdateStatusData {
	return p.data
}

// SetStatus sets the status of the presence
func (p *presence) SetStatus(status discordgo.Status) {
	p.data.Status = string(status)
}

// AddActivity adds an activity to the presence
func (p *presence) AddActivity(activity activity) {
	p.data.Activities = append(p.data.Activities, activity.data())
}

// activity interface for all activity types
type activity interface {
	data() discordgo.Activity
}

// customStatus holds data for a custom status
type customStatus struct {
	activity discordgo.Activity
}

// NewCustomStatus creates a new custom status
func NewCustomStatus(client *discordgo.Session) *customStatus {
	if client.State.User == nil {
		panic("Client is not logged in")
	}

	return &customStatus{
		activity: discordgo.Activity{
			Name: "Custom Status",
			Type: discordgo.ActivityTypeCustom,
		},
	}
}

// data returns the customStatus data
func (c *customStatus) data() discordgo.Activity {
	return c.activity
}

// SetState sets the state of the customStatus
func (c *customStatus) SetState(state string) {
	if len(state) > 128 {
		panic("State must be 128 characters or less")
	}

	c.activity.State = state
}

// SetEmoji sets the emoji of the customStatus
func (c *customStatus) SetEmoji(emoji string) {
	c.activity.Emoji = parseEmoji(emoji)
}

// baseActivity holds common activity data for richPresence and spotifyRichPresence
type baseActivity struct {
	activity discordgo.Activity
}

// data returns the baseActivity data
func (b *baseActivity) data() discordgo.Activity {
	if b.activity.Name == "" {
		panic("Name is required")
	}

	return b.activity
}

// SetApplicationID sets the application ID of the baseActivity
func (b *baseActivity) SetApplicationID(id string) {
	if !isValidDiscordId(id) {
		panic("Invalid application id")
	}

	b.activity.ApplicationID = id
}

// SetState sets the state of the baseActivity
func (b *baseActivity) SetState(state string) {
	if len(state) > 128 {
		panic("State must be 128 characters or less")
	}

	b.activity.State = state
}

// SetDetails sets the details of the baseActivity
func (b *baseActivity) SetDetails(details string) {
	if len(details) > 128 {
		panic("Details must be 128 characters or less")
	}

	b.activity.Details = details
}

// SetLargeImage sets the large image of the baseActivity
func (b *baseActivity) SetLargeImage(image string) {
	b.activity.Assets.LargeImageID = parseImage(image)
}

// SetLargeText sets the large text of the baseActivity
func (b *baseActivity) SetLargeText(text string) {
	if len(text) > 128 {
		panic("Large text must be 128 characters or less")
	}

	b.activity.Assets.LargeText = text
}

// SetSmallImage sets the small image of the baseActivity
func (b *baseActivity) SetSmallImage(image string) {
	b.activity.Assets.SmallImageID = parseImage(image)
}

// SetSmallText sets the small text of the baseActivity
func (b *baseActivity) SetSmallText(text string) {
	if len(text) > 128 {
		panic("Small text must be 128 characters or less")
	}

	b.activity.Assets.SmallText = text
}

// SetStartTimestamp sets the start timestamp of the baseActivity
func (b *baseActivity) SetStartTimestamp(timestamp time.Time) {
	b.activity.TimeStamps.StartTimestamp = timestamp.UnixMilli()
}

// SetEndTimestamp sets the end timestamp of the baseActivity
func (b *baseActivity) SetEndTimestamp(timestamp time.Time) {
	b.activity.TimeStamps.EndTimestamp = timestamp.UnixMilli()
}

// richPresence holds data for a rich presence
type richPresence struct {
	baseActivity
}

// NewRichPresence creates a new richPresence instance
func NewRichPresence(client *discordgo.Session) *richPresence {
	if client.State.User == nil {
		panic("Client is not logged in")
	}

	return &richPresence{}
}

// SetName sets the name of the richPresence
func (r *richPresence) SetName(name string) {
	if len(name) > 128 {
		panic("Name must be 128 characters or less")
	}

	r.activity.Name = name
}

// SetType sets the type of the richPresence
func (r *richPresence) SetType(activityType discordgo.ActivityType) {
	r.activity.Type = activityType
}

// SetURL sets the URL of the richPresence
func (r *richPresence) SetURL(url string) {
	if !isValidURL(url) {
		panic("Invalid URL")
	}

	r.activity.URL = url
}

// SetParty sets the party of the richPresence
func (r *richPresence) SetParty(id string, currentMembers, maxMembers int) {
	r.activity.Party = discordgo.Party{
		ID:   id,
		Size: []int{currentMembers, maxMembers},
	}
}

// AddButton adds a button to the richPresence
func (r *richPresence) AddButton(name, url string) {
	if len(r.activity.Buttons) >= 2 {
		panic("Rich presence can only have 2 buttons")
	}

	if name == "" || url == "" {
		panic("Button must have name and url")
	}

	if len(name) > 31 {
		panic("Button name must be less than 32 characters")
	}

	if !isValidURL(url) {
		panic("Invalid button URL")
	}

	r.activity.Buttons = append(r.activity.Buttons, name)
	r.activity.Metadata.ButtonUrls = append(r.activity.Metadata.ButtonUrls, url)
}

// spotifyRichPresence holds data for a Spotify rich presence
type spotifyRichPresence struct {
	baseActivity
}

// NewSpotifyRichPresence creates a new spotifyRichPresence instance
func NewSpotifyRichPresence(client *discordgo.Session) *spotifyRichPresence {
	if client.State.User == nil {
		panic("Client is not logged in")
	}

	return &spotifyRichPresence{
		baseActivity: baseActivity{
			activity: discordgo.Activity{
				ID:    "spotify:1",
				Name:  "Spotify",
				Type:  discordgo.ActivityTypeListening,
				Flags: 48,
				Party: discordgo.Party{
					ID: "spotify:" + client.State.User.ID,
				},
			},
		},
	}
}

// SetSongId sets the song ID of the spotifyRichPresence
func (s *spotifyRichPresence) SetSongId(id string) {
	if !isValidSpotifyId(id) {
		panic("Invalid song id")
	}

	s.activity.SyncId = id
}

// SetAlbumId sets the album ID of the spotifyRichPresence
func (s *spotifyRichPresence) SetAlbumId(id string) {
	if !isValidSpotifyId(id) {
		panic("Invalid album id")
	}

	s.activity.Metadata.AlbumId = id
	s.activity.Metadata.ContextUri = "spotify:album:" + id
}

// AddArtistId adds an artist ID to the spotifyRichPresence
func (s *spotifyRichPresence) AddArtistId(id string) {
	if !isValidSpotifyId(id) {
		panic("Invalid artist id")
	}

	s.activity.Metadata.ArtistIds = append(s.activity.Metadata.ArtistIds, id)
}

func ImageLinkToAsset(client *discordgo.Session, applicationId, imageLink string) string {
	if client.State.User == nil {
		panic("Client is not logged in")
	}

	if !isValidDiscordId(applicationId) {
		panic("Invalid application id")
	}

	if !isValidURL(imageLink) {
		panic("Invalid image URL")
	}

	assets, err := client.ExternalAssets(applicationId, []string{imageLink})
	if err != nil {
		panic(fmt.Sprintf("Error getting external assets: %s", err))
	}

	return assets[0].ExternalAssetPath
}

// isValidDiscordId checks if a string is a valid discord ID
func isValidDiscordId(id string) bool {
	return discordIdRegex.MatchString(id)
}

// isValidSpotifyId checks if a string is a valid Spotify ID
func isValidSpotifyId(id string) bool {
	return spotifyIdRegex.MatchString(id)
}

// isValidURL checks if a string is a valid URL
func isValidURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}

// parseEmoji parses an emoji from a string
func parseEmoji(text string) discordgo.SimpleEmoji {
	var emoji discordgo.SimpleEmoji

	if isValidDiscordId(text) {
		emoji.ID = text
		return emoji
	}

	if match := regexp.MustCompile(`<?(?:(a):)?(\w{2,32}):(\d{17,19})?>?`).FindStringSubmatch(text); match != nil {
		emoji.Name = match[2]
		emoji.ID = match[3]
		emoji.Animated = match[1] == "a"

		return emoji
	}

	emoji.Name = text
	return emoji
}

// parseImage parses an image from a string
func parseImage(image string) string {
	if isValidDiscordId(image) {
		return image
	}

	if strings.HasPrefix(image, "mp:") || strings.HasPrefix(image, "youtube:") || strings.HasPrefix(image, "spotify:") || strings.HasPrefix(image, "twitch:") {
		return image
	}

	if strings.HasPrefix(image, "external/") {
		return "mp:" + image
	}

	if isValidURL(image) {
		discordCdns := []string{
			"https://cdn.discordapp.com/",
			"http://cdn.discordapp.com/",
			"https://media.discordapp.net/",
			"http://media.discordapp.net/",
		}

		for _, cdn := range discordCdns {
			if strings.HasPrefix(image, cdn) {
				return strings.ReplaceAll(image, cdn, "mp:")
			}
		}
	}

	return image
}
