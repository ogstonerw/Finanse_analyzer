package events

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
)

type Service struct {
	assetsRepository *storage.AssetsRepository
	newsRepository   *storage.NewsItemsRepository
	eventsRepository *storage.EventsRepository
}

type Event struct {
	ID          string    `json:"id"`
	NewsItemID  string    `json:"news_item_id"`
	NewsTitle   string    `json:"news_title"`
	AssetID     *string   `json:"asset_id"`
	AssetTicker *string   `json:"asset_ticker"`
	AssetName   *string   `json:"asset_name"`
	EventType   string    `json:"event_type"`
	Summary     string    `json:"summary"`
	ExtractedAt time.Time `json:"extracted_at"`
}

type assetMatcher struct {
	id      string
	ticker  string
	name    string
	aliases []string
}

func NewService(store *storage.Postgres) *Service {
	return &Service{
		assetsRepository: storage.NewAssetsRepository(store),
		newsRepository:   storage.NewNewsItemsRepository(store),
		eventsRepository: storage.NewEventsRepository(store),
	}
}

func (s *Service) SyncFromNews(ctx context.Context) error {
	newsItems, err := s.newsRepository.List(ctx)
	if err != nil {
		return err
	}

	assets, err := s.assetsRepository.List(ctx)
	if err != nil {
		return err
	}

	matchers := buildAssetMatchers(assets)

	items := make([]storage.UpsertEventParams, 0, len(newsItems))
	for _, item := range newsItems {
		content := normalizeText(strings.Join([]string{item.Title, item.Summary, item.Body}, " "))
		assetID := matchAssetID(matchers, content)
		items = append(items, storage.UpsertEventParams{
			NewsItemID:  item.ID,
			AssetID:     assetID,
			EventType:   classifyEventType(content),
			Summary:     deriveEventSummary(item),
			ExtractedAt: time.Now().UTC(),
		})
	}

	return s.eventsRepository.UpsertBatch(ctx, items)
}

func (s *Service) List(ctx context.Context) ([]Event, error) {
	items, err := s.eventsRepository.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(items))
	for _, item := range items {
		result = append(result, Event{
			ID:          item.ID,
			NewsItemID:  item.NewsItemID,
			NewsTitle:   item.NewsTitle,
			AssetID:     nullStringToPointer(item.AssetID),
			AssetTicker: nullStringToPointer(item.AssetTicker),
			AssetName:   nullStringToPointer(item.AssetName),
			EventType:   item.EventType,
			Summary:     item.Summary,
			ExtractedAt: item.ExtractedAt,
		})
	}

	return result, nil
}

func buildAssetMatchers(assets []storage.AssetRecord) []assetMatcher {
	result := make([]assetMatcher, 0, len(assets))
	for _, asset := range assets {
		aliases := []string{
			normalizeText(asset.Ticker),
			normalizeText(asset.Name),
		}

		switch asset.Ticker {
		case "IMOEX":
			aliases = append(aliases, "MOEX RUSSIA INDEX")
		case "SBER":
			aliases = append(aliases, "SBERBANK")
		case "LKOH":
			aliases = append(aliases, "LUKOIL")
		case "GAZP":
			aliases = append(aliases, "GAZPROM")
		case "YDEX":
			aliases = append(aliases, "YANDEX")
		case "BRENT":
			aliases = append(aliases, "BRENT", "CRUDE OIL")
		case "NATGAS":
			aliases = append(aliases, "NATGAS", "NATURAL GAS", "HENRY HUB")
		}

		result = append(result, assetMatcher{
			id:      asset.ID,
			ticker:  asset.Ticker,
			name:    asset.Name,
			aliases: uniqueStrings(aliases),
		})
	}

	return result
}

func matchAssetID(matchers []assetMatcher, content string) *string {
	for _, matcher := range matchers {
		for _, alias := range matcher.aliases {
			if alias != "" && strings.Contains(content, alias) {
				return stringPointer(matcher.id)
			}
		}
	}

	return nil
}

func classifyEventType(content string) string {
	switch {
	case strings.Contains(content, "KEY RATE") && containsAny(content, "CUT", "REDUCE", "LOWER"):
		return "key_rate_cut"
	case strings.Contains(content, "KEY RATE") && containsAny(content, "KEEP", "HOLD", "UNCHANGED", "MAINTAIN"):
		return "key_rate_hold"
	case strings.Contains(content, "KEY RATE") && containsAny(content, "RAISE", "HIKE", "INCREASE"):
		return "key_rate_hike"
	case containsAny(content, "DIVIDEND", "DIVIDENDS"):
		return "dividend"
	case containsAny(content, "RESULTS", "EARNINGS", "REVENUE", "NET PROFIT"):
		return "financial_results"
	case containsAny(content, "SANCTION", "SANCTIONS"):
		return "sanctions"
	case containsAny(content, "BRENT", "CRUDE OIL"):
		return "commodity_oil"
	case containsAny(content, "NATURAL GAS", "HENRY HUB", "GAS PRICE"):
		return "commodity_gas"
	case containsAny(content, "MONETARY POLICY", "BANK OF RUSSIA", "INFLATION"):
		return "monetary_policy"
	default:
		return "general_news"
	}
}

func deriveEventSummary(item storage.NewsItemRecord) string {
	switch {
	case strings.TrimSpace(item.Summary) != "":
		return item.Summary
	case strings.TrimSpace(item.Title) != "":
		return item.Title
	default:
		return "Event extracted from normalized news item"
	}
}

func normalizeText(value string) string {
	value = strings.ToUpper(value)
	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		":", " ",
		";", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"/", " ",
		"\\", " ",
		"-", " ",
		"_", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)

	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func containsAny(value string, variants ...string) bool {
	for _, variant := range variants {
		if strings.Contains(value, variant) {
			return true
		}
	}

	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func stringPointer(value string) *string {
	return &value
}

func nullStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}
