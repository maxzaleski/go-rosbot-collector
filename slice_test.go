package rosbotcollector

import (
	"reflect"
	"testing"
)

func Test_contains(t *testing.T) {
	type args struct {
		s []Destination
		e Destination
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Does contains",
			args: args{
				s: []Destination{
					DestinationSalvaged,
					DestinationSold,
				},
				e: DestinationSalvaged,
			},
			want: true,
		},
		{
			name: "Does not contains",
			args: args{
				s: []Destination{},
				e: DestinationSalvaged,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.s, tt.args.e); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterLegendaries(t *testing.T) {
	type args struct {
		legendaries []*LegendaryItem
		config      ParserConfig
	}
	tests := []struct {
		name string
		args args
		want []*LegendaryItem
	}{
		{
			name: "Filter primals only",
			args: args{
				legendaries: []*LegendaryItem{
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityPrimal,
						Destination:  DestinationStashed,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityAncient,
						Destination:  DestinationSalvaged,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityNonAncient,
						Destination:  DestinationSalvaged,
						Stats:        "",
					},
				},
				config: ParserConfig{
					Destinations: []Destination{},
					RarityLevel:  RarityPrimal,
					Quality:      QualityNormal,
					Page:         1,
				},
			},
			want: []*LegendaryItem{
				{
					Name:         "unidentified",
					IsIdentified: false,
					Quality:      QualityNormal,
					Rarity:       RarityPrimal,
					Destination:  DestinationStashed,
					Stats:        "",
				},
			},
		},
		{
			name: "Filter Stashed only",
			args: args{
				legendaries: []*LegendaryItem{
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityPrimal,
						Destination:  DestinationStashed,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityAncient,
						Destination:  DestinationSalvaged,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityNonAncient,
						Destination:  DestinationSalvaged,
						Stats:        "",
					},
				},
				config: ParserConfig{
					Destinations: []Destination{DestinationStashed},
					RarityLevel:  RarityNonAncient,
					Quality:      QualityNormal,
					Page:         1,
				},
			},
			want: []*LegendaryItem{
				{
					Name:         "unidentified",
					IsIdentified: false,
					Quality:      QualityNormal,
					Rarity:       RarityPrimal,
					Destination:  DestinationStashed,
					Stats:        "",
				},
			},
		},
		{
			name: "Filter Stashed ancients & primals",
			args: args{
				legendaries: []*LegendaryItem{
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityPrimal,
						Destination:  DestinationStashed,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityAncient,
						Destination:  DestinationStashed,
						Stats:        "",
					},
					{
						Name:         "unidentified",
						IsIdentified: false,
						Quality:      QualityNormal,
						Rarity:       RarityNonAncient,
						Destination:  DestinationSalvaged,
						Stats:        "",
					},
				},
				config: ParserConfig{
					Destinations: []Destination{DestinationStashed},
					RarityLevel:  RarityAncient,
					Quality:      QualityNormal,
					Page:         1,
				},
			},
			want: []*LegendaryItem{
				{
					Name:         "unidentified",
					IsIdentified: false,
					Quality:      QualityNormal,
					Rarity:       RarityPrimal,
					Destination:  DestinationStashed,
					Stats:        "",
				},
				{
					Name:         "unidentified",
					IsIdentified: false,
					Quality:      QualityNormal,
					Rarity:       RarityAncient,
					Destination:  DestinationStashed,
					Stats:        "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterItems(tt.args.legendaries, &tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterItems() = %v, want %v", got, tt.want)
			}
		})
	}
}
