package rosbotcollector

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseDestination(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want Destination
	}{
		{
			name: "Salvaged titlecase",
			args: args{raw: "botname: Salvaged item name"},
			want: DestinationSalvaged,
		},
		{
			name: "Stashed titlecase",
			args: args{raw: "botname: Stashed item name"},
			want: DestinationStashed,
		},
		{
			name: "Sold titlecase",
			args: args{raw: "botname: Sold item name"},
			want: DestinationSold,
		},
		{
			name: "Salvaged lowercase",
			args: args{raw: "botname: salvaged item name"},
			want: DestinationSalvaged,
		},
		{
			name: "Stashed lowercase",
			args: args{raw: "botname: stashed item name"},
			want: DestinationStashed,
		},
		{
			name: "Sold lowercase",
			args: args{raw: "botname: sold item name"},
			want: DestinationSold,
		},
		{
			name: "Unknown",
			args: args{raw: "botname: Test item name"},
			want: DestinationUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDestination(tt.args.raw); got != tt.want {
				t.Errorf("parseDestination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseTimestamp(t *testing.T) {
	want, _ := time.Parse("02/01/2006 15:04", "05/10/2001 14:55")
	wantInvalid, _ := time.Parse("02/01/2006 15:04", "05/10/20 14:55")

	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "valid input format",
			args: args{raw: "05/10/2001 - 14:55"},
			want: want,
		},
		{
			name: "invalid input format test 1",
			args: args{raw: "05/10/20 - 14:55"},
			want: wantInvalid,
		},
		{
			name: "invalid input format test 2",
			args: args{raw: "05/10/20 - 14:55:50"},
			want: wantInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTimestamp(tt.args.raw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseItemType(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want Quality
	}{
		{
			name: "Normal",
			args: args{raw: "Text-Legendary"},
			want: QualityNormal,
		},
		{
			name: "Set",
			args: args{raw: "Text-Set"},
			want: QualitySet,
		},
		{
			name: "Other",
			args: args{raw: "Text-Rare"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseItemQuality(tt.args.raw); got != tt.want {
				t.Errorf("parseItemQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createSearchSegment(t *testing.T) {
	type args struct {
		config ParserConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Multiple destinations",
			args: args{
				config: ParserConfig{
					Destinations: []Destination{
						DestinationSalvaged,
						DestinationSold,
					},
					RarityLevel: RarityNonAncient,
					Quality:     QualityNormal,
					Page:        1,
				},
			},
			want: "/?item_destination=All&ancient=0&item_quality=3&page=1",
		},
		{
			name: "Single destination",
			args: args{
				config: ParserConfig{
					Destinations: []Destination{DestinationSalvaged},
					RarityLevel:  RarityNonAncient,
					Quality:      QualityNormal,
					Page:         1,
				},
			},
			want: "/?item_destination=2&ancient=0&item_quality=3&page=1",
		},
		{
			name: "Rarity ancient",
			args: args{
				config: ParserConfig{
					Destinations: []Destination{DestinationSalvaged},
					RarityLevel:  RarityAncient,
					Quality:      QualityNormal,
					Page:         1,
				},
			},
			want: "/?item_destination=2&ancient=1&item_quality=3&page=1",
		},
		{
			name: "Quality set",
			args: args{
				config: ParserConfig{
					Destinations: []Destination{DestinationSalvaged},
					RarityLevel:  RarityNonAncient,
					Quality:      QualitySet,
					Page:         1,
				},
			},
			want: "/?item_destination=2&ancient=0&item_quality=4&page=1",
		},
		{
			name: "All",
			args: args{
				config: ParserConfig{
					Destinations: []Destination{},
					RarityLevel:  RarityNonAncient,
					Quality:      QualityAll,
					Page:         1,
				},
			},
			want: "/?item_destination=All&ancient=0&item_quality=All&page=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assignSearchParams(&tt.args.config); got != tt.want {
				t.Errorf("assignSearchParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseItemName(t *testing.T) {
	type args struct {
		raw    string
		rarity Rarity
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "name with rarity",
			args: args{
				raw:    "[Primal] test name",
				rarity: RarityPrimal,
			},
			want: "test name",
		},
		{
			name: "name with no rarity",
			args: args{
				raw:    "test name",
				rarity: RarityNonAncient,
			},
			want: "test name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseItemName(tt.args.raw, tt.args.rarity); got != tt.want {
				t.Errorf("parseItemName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseItemRarity(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want Rarity
	}{
		{
			name: "Primal",
			args: args{
				raw: "[Primal] name",
			},
			want: RarityPrimal,
		},
		{
			name: "Ancient",
			args: args{
				raw: "[Ancient] name",
			},
			want: RarityAncient,
		},
		{
			name: "Non-Ancient",
			args: args{
				raw: "name",
			},
			want: RarityNonAncient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseItemRarity(tt.args.raw); got != tt.want {
				t.Errorf("parseItemRarity() = %v, want %v", got, tt.want)
			}
		})
	}
}
