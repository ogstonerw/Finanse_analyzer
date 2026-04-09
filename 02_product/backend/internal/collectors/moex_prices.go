package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	moexISSBaseURL      = "https://iss.moex.com/iss"
	moexDailyInterval   = 24
	moexPageSize        = 500
	moexDateParamLayout = "2006-01-02"
	moexDateTimeLayout  = "2006-01-02 15:04:05"
)

type MOEXInstrument struct {
	Ticker   string
	Engine   string
	Market   string
	Board    string
	Security string
}

type MOEXCandle struct {
	Time   time.Time
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
}

type MOEXCollector struct {
	client      *http.Client
	baseURL     string
	instruments []MOEXInstrument
	byTicker    map[string]MOEXInstrument
}

type moexCandlesResponse struct {
	Candles struct {
		Columns []string `json:"columns"`
		Data    [][]any  `json:"data"`
	} `json:"candles"`
}

func NewMOEXCollector(client *http.Client) *MOEXCollector {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	instruments := []MOEXInstrument{
		{Ticker: "IMOEX", Engine: "stock", Market: "index", Board: "SNDX", Security: "IMOEX"},
		{Ticker: "SBER", Engine: "stock", Market: "shares", Board: "TQBR", Security: "SBER"},
		{Ticker: "LKOH", Engine: "stock", Market: "shares", Board: "TQBR", Security: "LKOH"},
		{Ticker: "GAZP", Engine: "stock", Market: "shares", Board: "TQBR", Security: "GAZP"},
		{Ticker: "YDEX", Engine: "stock", Market: "shares", Board: "TQBR", Security: "YDEX"},
	}

	byTicker := make(map[string]MOEXInstrument, len(instruments))
	for _, item := range instruments {
		byTicker[item.Ticker] = item
	}

	return &MOEXCollector{
		client:      client,
		baseURL:     moexISSBaseURL,
		instruments: instruments,
		byTicker:    byTicker,
	}
}

func (c *MOEXCollector) SupportedTickers() []string {
	result := make([]string, 0, len(c.instruments))
	for _, item := range c.instruments {
		result = append(result, item.Ticker)
	}

	return result
}

func (c *MOEXCollector) FetchDailyCandles(ctx context.Context, ticker string, from, till time.Time) ([]MOEXCandle, error) {
	instrument, ok := c.byTicker[strings.ToUpper(strings.TrimSpace(ticker))]
	if !ok {
		return nil, fmt.Errorf("moex instrument mapping is not configured for ticker %s", ticker)
	}

	if till.IsZero() {
		till = time.Now().UTC()
	}

	start := 0
	items := make([]MOEXCandle, 0)
	for {
		page, err := c.fetchDailyCandlesPage(ctx, instrument, from, till, start)
		if err != nil {
			return nil, err
		}

		items = append(items, page...)
		if len(page) < moexPageSize {
			break
		}

		start += len(page)
	}

	return items, nil
}

func (c *MOEXCollector) fetchDailyCandlesPage(ctx context.Context, instrument MOEXInstrument, from, till time.Time, start int) ([]MOEXCandle, error) {
	endpoint, err := c.buildCandlesURL(instrument, from, till, start)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create moex request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request moex candles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected moex status: %s", resp.Status)
	}

	var payload moexCandlesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode moex candles response: %w", err)
	}

	return decodeMOEXCandles(payload)
}

func (c *MOEXCollector) buildCandlesURL(instrument MOEXInstrument, from, till time.Time, start int) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse moex base url: %w", err)
	}

	u.Path = fmt.Sprintf(
		"/iss/engines/%s/markets/%s/boards/%s/securities/%s/candles.json",
		instrument.Engine,
		instrument.Market,
		instrument.Board,
		instrument.Security,
	)

	query := u.Query()
	query.Set("interval", strconv.Itoa(moexDailyInterval))
	query.Set("from", from.UTC().Format(moexDateParamLayout))
	query.Set("till", till.UTC().Format(moexDateParamLayout))
	query.Set("start", strconv.Itoa(start))
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func decodeMOEXCandles(payload moexCandlesResponse) ([]MOEXCandle, error) {
	columnIndex := make(map[string]int, len(payload.Candles.Columns))
	for index, column := range payload.Candles.Columns {
		columnIndex[column] = index
	}

	requiredColumns := []string{"open", "close", "high", "low", "volume", "begin"}
	for _, column := range requiredColumns {
		if _, ok := columnIndex[column]; !ok {
			return nil, fmt.Errorf("moex candles response is missing column %s", column)
		}
	}

	items := make([]MOEXCandle, 0, len(payload.Candles.Data))
	for _, row := range payload.Candles.Data {
		item, err := decodeMOEXCandleRow(row, columnIndex)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func decodeMOEXCandleRow(row []any, columnIndex map[string]int) (MOEXCandle, error) {
	openValue, err := valueAsFloat64(row[columnIndex["open"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode open price: %w", err)
	}

	closeValue, err := valueAsFloat64(row[columnIndex["close"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode close price: %w", err)
	}

	highValue, err := valueAsFloat64(row[columnIndex["high"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode high price: %w", err)
	}

	lowValue, err := valueAsFloat64(row[columnIndex["low"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode low price: %w", err)
	}

	volumeValue, err := valueAsFloat64(row[columnIndex["volume"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode volume: %w", err)
	}

	candleTime, err := valueAsTime(row[columnIndex["begin"]])
	if err != nil {
		return MOEXCandle{}, fmt.Errorf("decode candle time: %w", err)
	}

	return MOEXCandle{
		Time:   candleTime,
		Open:   openValue,
		Close:  closeValue,
		High:   highValue,
		Low:    lowValue,
		Volume: volumeValue,
	}, nil
}

func valueAsFloat64(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", value)
	}
}

func valueAsTime(value any) (time.Time, error) {
	raw, ok := value.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("unsupported time type %T", value)
	}

	parsed, err := time.ParseInLocation(moexDateTimeLayout, raw, time.UTC)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}
