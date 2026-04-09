package collectors

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	cbrMonetaryPolicyURL = "https://www.cbr.ru/eng/dkp/mp_dec/"
	cbrArticleRootURL    = "https://www.cbr.ru"
)

var (
	cbrDayBlockPattern      = regexp.MustCompile(`(?s)<div class="previews_day".*?</div>\s*</div>`)
	cbrDayDatePattern       = regexp.MustCompile(`previews_day-date">\s*([^<]+?)\s*</div>`)
	cbrPreviewItemPattern   = regexp.MustCompile(`(?s)<div class="previews_item">.*?<div class="previews_item-time">\s*([^<]+?)\s*</div>.*?<div class="previews_item-title"><a href="([^"]+)">([^<]+)</a>`)
	cbrParagraphPattern     = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	cbrTagPattern           = regexp.MustCompile(`(?s)<[^>]+>`)
	cbrWhitespacePattern    = regexp.MustCompile(`\s+`)
	cbrTimestampTextPattern = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}\s+\d{2}\.\d{2}\.\d{2}$`)
)

type CollectedNewsItem struct {
	ExternalID  string
	Title       string
	Summary     string
	Body        string
	PublishedAt time.Time
	CollectedAt time.Time
	URL         string
}

type NewsCollector interface {
	SourceBaseURL() string
	CollectLatest(ctx context.Context) ([]CollectedNewsItem, error)
}

type CBRMonetaryPolicyCollector struct {
	client      *http.Client
	sourceURL   string
	articleRoot string
}

func NewCBRMonetaryPolicyCollector(client *http.Client) *CBRMonetaryPolicyCollector {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &CBRMonetaryPolicyCollector{
		client:      client,
		sourceURL:   cbrMonetaryPolicyURL,
		articleRoot: cbrArticleRootURL,
	}
}

func (c *CBRMonetaryPolicyCollector) SourceBaseURL() string {
	return c.sourceURL
}

func (c *CBRMonetaryPolicyCollector) CollectLatest(ctx context.Context) ([]CollectedNewsItem, error) {
	page, err := c.fetchText(ctx, c.sourceURL)
	if err != nil {
		return nil, err
	}

	tabContent := extractCBRPrimaryTab(page)
	dayBlocks := cbrDayBlockPattern.FindAllString(tabContent, -1)
	if len(dayBlocks) == 0 {
		return nil, fmt.Errorf("parse cbr news list: no preview blocks found")
	}

	items := make([]CollectedNewsItem, 0)
	collectedAt := time.Now().UTC()
	for _, block := range dayBlocks {
		dayText, ok := firstMatch(cbrDayDatePattern, block)
		if !ok {
			continue
		}

		previews := cbrPreviewItemPattern.FindAllStringSubmatch(block, -1)
		for _, preview := range previews {
			publishedAt, err := parseCBRPublishedAt(dayText, preview[1])
			if err != nil {
				return nil, fmt.Errorf("parse cbr published time: %w", err)
			}

			itemURL, err := resolveURL(c.articleRoot, preview[2])
			if err != nil {
				return nil, fmt.Errorf("resolve cbr article url: %w", err)
			}

			body, summary, err := c.fetchArticleBody(ctx, itemURL)
			if err != nil {
				return nil, err
			}

			items = append(items, CollectedNewsItem{
				ExternalID:  deriveExternalID(itemURL),
				Title:       cleanText(preview[3]),
				Summary:     summary,
				Body:        body,
				PublishedAt: publishedAt,
				CollectedAt: collectedAt,
				URL:         itemURL,
			})
		}
	}

	return items, nil
}

func (c *CBRMonetaryPolicyCollector) fetchArticleBody(ctx context.Context, articleURL string) (string, string, error) {
	page, err := c.fetchText(ctx, articleURL)
	if err != nil {
		return "", "", err
	}

	articleContent := extractCBRArticleContent(page)
	paragraphMatches := cbrParagraphPattern.FindAllStringSubmatch(articleContent, -1)
	paragraphs := make([]string, 0, len(paragraphMatches))
	for _, match := range paragraphMatches {
		text := cleanText(match[1])
		if text == "" {
			continue
		}
		if strings.Contains(text, "The reference to the Press Service is mandatory") {
			continue
		}
		if cbrTimestampTextPattern.MatchString(text) {
			continue
		}

		paragraphs = append(paragraphs, text)
	}

	if len(paragraphs) == 0 {
		return "", "", nil
	}

	return strings.Join(paragraphs, "\n\n"), paragraphs[0], nil
}

func (c *CBRMonetaryPolicyCollector) fetchText(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("create cbr request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request cbr page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected cbr status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read cbr response body: %w", err)
	}

	return string(body), nil
}

func extractCBRPrimaryTab(page string) string {
	start := strings.Index(page, `id="tab_content_t1"`)
	if start == -1 {
		return page
	}

	content := page[start:]
	end := strings.Index(content, `id="tab_content_t2"`)
	if end == -1 {
		return content
	}

	return content[:end]
}

func extractCBRArticleContent(page string) string {
	start := strings.Index(page, `class="landing-text"`)
	if start == -1 {
		return page
	}

	content := page[start:]
	for _, marker := range []string{`class="page-share"`, `class="article__tags"`, `class="versions"`, `</article>`} {
		if end := strings.Index(content, marker); end != -1 {
			return content[:end]
		}
	}

	return content
}

func parseCBRPublishedAt(dayText, timeText string) (time.Time, error) {
	location := time.FixedZone("MSK", 3*60*60)
	combined := fmt.Sprintf("%s %s", cleanText(dayText), cleanText(timeText))

	parsed, err := time.ParseInLocation("2 January 2006 15:04", combined, location)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}

func resolveURL(baseURL, href string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	target, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return "", err
	}

	return base.ResolveReference(target).String(), nil
}

func deriveExternalID(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if fileValue := strings.TrimSpace(parsed.Query().Get("file")); fileValue != "" {
		return fileValue
	}

	return strings.Trim(parsed.Path, "/")
}

func firstMatch(pattern *regexp.Regexp, value string) (string, bool) {
	match := pattern.FindStringSubmatch(value)
	if len(match) < 2 {
		return "", false
	}

	return match[1], true
}

func cleanText(raw string) string {
	withoutTags := cbrTagPattern.ReplaceAllString(raw, " ")
	unescaped := html.UnescapeString(withoutTags)
	return strings.TrimSpace(cbrWhitespacePattern.ReplaceAllString(unescaped, " "))
}
