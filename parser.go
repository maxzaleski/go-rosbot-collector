package rosbotcollector

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type (
	Parser interface {
		// Parse parses server updates from the '/bot-activity' page.
		Parse(ctx context.Context) ([]*ServerUpdate, error)
	}

	parser struct {
		config      *ParserConfig
		httpService HTTPService
	}
)

func newParser(c *ParserConfig, s HTTPService) Parser {
	return &parser{
		config:      c,
		httpService: s,
	}
}

func (p *parser) Parse(ctx context.Context) ([]*ServerUpdate, error) {
	body, err := p.httpService.GetActivity(assignSearchParams(p.config))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	rawUpdates := doc.Find("div.timeline-item")

	// Every server update is parsed concurrently.
	// For every update (u • typically 2-4) there are u * 3 go routines spawned which
	// concurrently parse the legendary rawUpdates.
	updateChan := make(chan *ServerUpdate, rawUpdates.Length())
	wg := &sync.WaitGroup{}
	wg.Add(rawUpdates.Length())

	rawUpdates.Each(func(_ int, s *goquery.Selection) {
		go parseUpdate(ctx, wg, updateChan, s, p.config)
	})

	wg.Wait()
	close(updateChan)

	parsedUpdates := make([]*ServerUpdate, 0, rawUpdates.Length())
	for u := range updateChan {
		parsedUpdates = append(parsedUpdates, u)
	}

	// Since every update is parsed concurrently, entries may not be sorted.
	sort.Slice(parsedUpdates, func(i, j int) bool {
		return parsedUpdates[i].ServerTimestamp.Before(parsedUpdates[j].ServerTimestamp)
	})
	return parsedUpdates, nil
}

func parseUpdate(
	ctx context.Context,
	wg *sync.WaitGroup,
	out chan<- *ServerUpdate,
	s *goquery.Selection,
	config *ParserConfig,
) {
	defer wg.Done()

	items := s.Find("p.m-b-xs")
	itemsChan := make(chan *LegendaryItem, items.Length())

	jobs := make(chan *goquery.Selection)
	for i := 0; i <= 3; i++ {
		go parseLegendaryItemWorker(ctx, jobs, itemsChan)
	}
	items.Each(func(_ int, s *goquery.Selection) { jobs <- s })
	close(jobs)

	legendaryItems := make([]*LegendaryItem, 0, items.Length())
	for item := range itemsChan {
		legendaryItems = append(legendaryItems, item)
	}
	close(itemsChan)

	out <- &ServerUpdate{
		ServerTimestamp: parseTimestamp(s.Find("div.date").Text()),
		Items:           filterItems(legendaryItems, config),
	}
}

func parseLegendaryItemWorker(
	ctx context.Context,
	jobs <-chan *goquery.Selection,
	out chan<- *LegendaryItem,
) {
	/*
		Example of an identified legendary

		<p class="m-b-xs">
			TestDiablo3Name: Salvaged
			<span
				data-toggle="popover"
				data-placement="right"
				data-trigger="hover"
				data-html="true"
				data-title="tyrael&#39;s might"
				data-content="
				Armor&lt;br /&gt;
				736&lt;br /&gt;
				Primary&lt;br /&gt;
				+474 Dexterity&lt;br /&gt;
				+430 Vitality&lt;br /&gt;
				+95 Resistance to All Elements&lt;br /&gt;
				Secondary&lt;br /&gt;
				+19% Damage to Demons&lt;br /&gt;
				+5840 Life after Each Kill&lt;br /&gt;
				Ignores Durability Loss&lt;br /&gt;
				1 Socket(s)"
				class="text-LegendaryItem"
				data-original-title=""
				title="">tyrael's might
			</span>
		</p>
	*/
	for j := range jobs {
		span := j.Find("span")

		// These attributes are always present; presence feedback is ignored.
		rawStats, _ := span.Attr("data-content")
		rawClass, _ := span.Attr("class")
		rawSpanText := strings.TrimSpace(span.Text())

		q := parseItemQuality(rawClass)
		// Quality is below "legendary".
		if q == "" {
			// Ignore for now.
			// Lower quality items could be added later on.
			continue
		}

		r := parseItemRarity(rawSpanText)
		n := parseItemName(rawSpanText, r)

		out <- &LegendaryItem{
			Name:         n,
			Rarity:       r,
			Quality:      q,
			IsIdentified: n == "unidentified",
			Destination:  parseDestination(j.Text()),
			Stats:        parseItemStats(rawStats),
		}
	}
}

var timestampRegex = regexp.MustCompile(`\d{2}/\d{2}/\d{4}\s-\s\d{2}:\d{2}`)

func parseTimestamp(raw string) time.Time {
	// If an error were to occur, a bad input from the website itself most likely,
	// the parse function will just return 0001-01-01 00:00:00 +0000 UTC which is acceptable.

	parsed := timestampRegex.FindString(strings.TrimSpace(raw))
	parsed = strings.ReplaceAll(parsed, " -", "")

	t, _ := time.Parse("02/01/2006 15:04", parsed)
	return t
}

var destinationRegex = regexp.MustCompile(`:\s([a-zA-Z]+)`)

func parseDestination(raw string) Destination {
	switch strings.ToLower(destinationRegex.FindStringSubmatch(raw)[1]) {
	case "salvaged":
		return DestinationSalvaged
	case "stashed":
		return DestinationStashed
	case "sold":
		return DestinationSold
	default:
		return DestinationUnknown
	}
}

var whiteSpaceRegex = regexp.MustCompile(`^ *`)

func parseItemStats(raw string) string {
	// Not enough test data implement precise logic.
	// More prudent to let the end-user decide how the data should be handled and presented.
	//
	// The package will handle sanitising the parsed data,
	// and return it under the form foo\nbar\nfor\nbar...

	// Remove redundant message.
	result := strings.ReplaceAll(raw, "This item cannot be equipped until it is identified.", "")

	// Remove redundant break tag.
	result = strings.ReplaceAll(result, "<br />", "")

	// Remove redundant whitespace at start of line.
	// Awkward implementation as we can't match any of the wanted whitespaces until each line is
	// parsed individually.
	//
	// We could use a scanner to achieve a similar result. However, I have no immediate plans of
	// benchmarking either solution.
	s := strings.Split(result, "\n")
	for i, line := range s {
		s[i] = whiteSpaceRegex.ReplaceAllString(line, "")
	}

	return strings.Join(s, "\n")
}

func parseItemName(raw string, rarity Rarity) string {
	name := raw
	if rarity != RarityNonAncient {
		name = strings.Split(raw, "] ")[1]
	}
	return name
}

var rarityRegex = regexp.MustCompile(`\[[a-zA-Z]+]`)

func parseItemRarity(raw string) Rarity {
	// Can also be done through the class name.
	// However, there is no distinction between "ancient" and "normal" items.
	switch strings.ToLower(rarityRegex.FindString(raw)) {
	case "[ancient]":
		return RarityAncient
	case "[primal]":
		return RarityPrimal
	default:
		return RarityNonAncient
	}
}

func parseItemQuality(raw string) Quality {
	switch strings.ToLower(raw) {
	case "text-legendary":
		return QualityNormal
	case "text-set":
		return QualitySet
	default:
		return ""
	}
}

func assignSearchParams(config *ParserConfig) string {
	var (
		destination,
		rarity,
		quality string
	)

	if len(config.Destinations) == 1 {
		switch config.Destinations[0] {
		case DestinationStashed:
			destination = "1"
			break
		case DestinationSalvaged:
			destination = "2"
			break
		case DestinationSold:
			destination = "4"
			break
		}
	} else {
		destination = "All"
	}

	switch config.RarityLevel {
	case RarityPrimal:
	case RarityAncient:
		rarity = "1"
	default:
		rarity = "0"
	}

	switch config.Quality {
	case QualityNormal:
		quality = "3"
		break
	case QualitySet:
		quality = "4"
		break
	default:
		quality = "All"
		break
	}

	return fmt.Sprintf(
		"/?item_destination=%s&ancient=%s&item_quality=%s&page=%d",
		destination, rarity, quality, config.Page,
	)
}

// ParserConfig is the parsing configuration
type ParserConfig struct {
	Destinations []Destination
	RarityLevel  Rarity
	Quality      Quality
	Page         int8
}

// NewParseConfig returns a new instance of `rosbotcollector.ParserConfig` with the default values.
func NewParseConfig() *ParserConfig {
	return &ParserConfig{
		Destinations: []Destination{},
		RarityLevel:  RarityNonAncient,
		Quality:      QualityAll,
		Page:         1,
	}
}

// ServerUpdate is a Ros-Bot server update.
type ServerUpdate struct {
	Items           []*LegendaryItem `json:"legendaries"`
	ServerTimestamp time.Time        `json:"server_timestamp"`
}

// LegendaryItem is a Diablo III legendary item.
type LegendaryItem struct {
	Name         string      `json:"name"`
	Quality      Quality     `json:"type"`
	Rarity       Rarity      `json:"rarity"`
	Destination  Destination `json:"destination"`
	IsIdentified bool        `json:"is_identified"`
	Stats        string      `json:"stats"`
}

// Destination is where the bot placed the item upon collection of it.
type Destination string

const (
	DestinationStashed  Destination = "STASHED"
	DestinationSalvaged Destination = "SALVAGED"
	DestinationSold     Destination = "SOLD"
	DestinationUnknown  Destination = "UNKNOWN"
)

// Rarity is the item's rarity.
type Rarity string

const (
	RarityPrimal     Rarity = "PRIMAL"
	RarityAncient    Rarity = "ANCIENT"
	RarityNonAncient Rarity = "NON-ANCIENT"
)

// Quality is the item's quality.
type Quality string

const (
	QualityAll    Quality = "*"
	QualityNormal Quality = "NORMAL"
	QualitySet    Quality = "SET"
)
