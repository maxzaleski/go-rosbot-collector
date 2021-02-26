package rosbotcollector

func contains(s []Destination, e Destination) bool {
	for _, _e := range s {
		if _e == e {
			return true
		}
	}
	return false
}

func filter(s []*LegendaryItem, condition func(e *LegendaryItem) bool) (items []*LegendaryItem) {
	for _, e := range s {
		if condition(e) {
			items = append(items, e)
		}
	}
	return
}

func filterItems(legendaries []*LegendaryItem, config *ParserConfig) []*LegendaryItem {
	return filter(legendaries, func(item *LegendaryItem) bool {
		if (config.RarityLevel == RarityAncient && item.Rarity == RarityNonAncient) ||
			(config.RarityLevel == RarityPrimal && item.Rarity != RarityPrimal) {
			return false
		}

		if config.Quality != QualityAll && item.Quality != config.Quality {
			return false
		}

		if len(config.Destinations) != 0 {
			return contains(config.Destinations, item.Destination)
		}

		return true
	})
}
