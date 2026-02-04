package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	apiURL   = "https://developers.events/all-cfps.json"
	cacheTTL = 1 * time.Hour
)

type Conference struct {
	Name      string   `json:"name"`
	Date      []int64  `json:"date"`
	Hyperlink string   `json:"hyperlink"`
	Status    string   `json:"status"`
	Location  string   `json:"location"`
}

type CFP struct {
	Link      string      `json:"link"`
	Until     string      `json:"until"`
	UntilDate int64       `json:"untilDate"`
	Conf      *Conference `json:"conf"`
}

type CFPResult struct {
	Conference        string `json:"conference"`
	Location          string `json:"location"`
	CFPDeadline       string `json:"cfpDeadline"`
	DaysRemaining     int    `json:"daysRemaining"`
	ConferenceDate    string `json:"conferenceDate"`
	CFPLink           string `json:"cfpLink"`
	ConferenceWebsite string `json:"conferenceWebsite"`
	Status            string `json:"status"`
}

type CFPCache struct {
	data      []CFP
	timestamp time.Time
}

var cache *CFPCache

// Tool argument types
type ListOpenCFPsArgs struct {
	Limit int `json:"limit" jsonschema:"Maximum number of results to return,default=20"`
}

type SearchByKeywordArgs struct {
	Keywords []string `json:"keywords" jsonschema:"List of keywords to search for in conference names and CFP links,required"`
	Limit    int      `json:"limit" jsonschema:"Maximum number of results to return,default=20"`
}

type FindClosingCFPsArgs struct {
	DaysAhead int      `json:"daysAhead" jsonschema:"Number of days to look ahead,default=7"`
	Keywords  []string `json:"keywords" jsonschema:"Optional keywords to search for in conference names"`
}

type SearchByLocationArgs struct {
	Location string `json:"location" jsonschema:"Location to search for (e.g. France, USA, Online),required"`
	Limit    int    `json:"limit" jsonschema:"Maximum number of results to return,default=20"`
}

// Response types
type ListOpenCFPsResponse struct {
	TotalOpen int         `json:"totalOpen"`
	Returned  int         `json:"returned"`
	CFPs      []CFPResult `json:"cfps"`
}

type SearchByKeywordResponse struct {
	SearchKeywords []string    `json:"searchKeywords"`
	TotalMatches   int         `json:"totalMatches"`
	CFPs           []CFPResult `json:"cfps"`
}

type FindClosingCFPsResponse struct {
	DaysAhead      int         `json:"daysAhead"`
	FilterKeywords string      `json:"filterKeywords"`
	UrgentCFPs     int         `json:"urgentCFPs"`
	CFPs           []CFPResult `json:"cfps"`
}

type SearchByLocationResponse struct {
	SearchLocation string      `json:"searchLocation"`
	TotalMatches   int         `json:"totalMatches"`
	CFPs           []CFPResult `json:"cfps"`
}

func fetchCFPs() ([]CFP, error) {
	now := time.Now()
	if cache != nil && now.Sub(cache.timestamp) < cacheTTL {
		return cache.data, nil
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CFPs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cfps []CFP
	if err := json.Unmarshal(body, &cfps); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	cache = &CFPCache{
		data:      cfps,
		timestamp: now,
	}

	return cfps, nil
}

func getOpenCFPs() ([]CFP, error) {
	allCFPs, err := fetchCFPs()
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	var openCFPs []CFP
	for _, cfp := range allCFPs {
		if cfp.UntilDate > now {
			openCFPs = append(openCFPs, cfp)
		}
	}

	return openCFPs, nil
}

func toCFPResult(cfp CFP) CFPResult {
	now := time.Now().UnixMilli()
	diffMs := cfp.UntilDate - now
	daysRemaining := int((diffMs + 86399999) / 86400000)

	confDate := "Unknown"
	if cfp.Conf != nil && len(cfp.Conf.Date) > 0 {
		t := time.UnixMilli(cfp.Conf.Date[0])
		confDate = t.Format("2006-01-02")
	}

	conference := "Unknown"
	location := "Unknown"
	conferenceWebsite := ""
	status := "unknown"

	if cfp.Conf != nil {
		conference = cfp.Conf.Name
		location = cfp.Conf.Location
		conferenceWebsite = cfp.Conf.Hyperlink
		status = cfp.Conf.Status
	}

	return CFPResult{
		Conference:        conference,
		Location:          location,
		CFPDeadline:       cfp.Until,
		DaysRemaining:     daysRemaining,
		ConferenceDate:    confDate,
		CFPLink:           cfp.Link,
		ConferenceWebsite: conferenceWebsite,
		Status:            status,
	}
}

func matchesKeywords(cfp CFP, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}

	searchText := strings.ToLower(cfp.Link)
	if cfp.Conf != nil {
		searchText += " " + strings.ToLower(cfp.Conf.Name)
		searchText += " " + strings.ToLower(cfp.Conf.Location)
	}

	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func matchesLocation(cfp CFP, location string) bool {
	if location == "" {
		return true
	}
	if cfp.Conf == nil {
		return false
	}
	return strings.Contains(strings.ToLower(cfp.Conf.Location), strings.ToLower(location))
}

func createServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "developer-events-server",
		Version: "1.0.0",
	}, nil)

	// Tool 1: List open CFPs
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_open_cfps",
		Description: "List all currently open Call for Papers (CFPs) for developer conferences",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListOpenCFPsArgs) (*mcp.CallToolResult, ListOpenCFPsResponse, error) {
		limit := args.Limit
		if limit == 0 {
			limit = 20
		}

		openCFPs, err := getOpenCFPs()
		if err != nil {
			return nil, ListOpenCFPsResponse{}, err
		}

		totalOpen := len(openCFPs)

		// Sort by deadline
		sort.Slice(openCFPs, func(i, j int) bool {
			return openCFPs[i].UntilDate < openCFPs[j].UntilDate
		})

		// Limit results
		if len(openCFPs) > limit {
			openCFPs = openCFPs[:limit]
		}

		results := make([]CFPResult, len(openCFPs))
		for i, cfp := range openCFPs {
			results[i] = toCFPResult(cfp)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d open CFPs, returning %d", totalOpen, len(results))},
			},
		}, ListOpenCFPsResponse{
			TotalOpen: totalOpen,
			Returned:  len(results),
			CFPs:      results,
		}, nil
	})

	// Tool 2: Search by keywords
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_cfps_by_keyword",
		Description: "Search open CFPs by keywords in conference names and CFP links (e.g., 'java', 'python', 'AI', 'cloud', 'kubernetes')",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchByKeywordArgs) (*mcp.CallToolResult, SearchByKeywordResponse, error) {
		limit := args.Limit
		if limit == 0 {
			limit = 20
		}

		openCFPs, err := getOpenCFPs()
		if err != nil {
			return nil, SearchByKeywordResponse{}, err
		}

		// Filter by keywords
		var filtered []CFP
		for _, cfp := range openCFPs {
			if matchesKeywords(cfp, args.Keywords) {
				filtered = append(filtered, cfp)
			}
		}

		// Sort by deadline
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].UntilDate < filtered[j].UntilDate
		})

		// Limit results
		if len(filtered) > limit {
			filtered = filtered[:limit]
		}

		results := make([]CFPResult, len(filtered))
		for i, cfp := range filtered {
			results[i] = toCFPResult(cfp)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d CFPs matching keywords %v", len(results), args.Keywords)},
			},
		}, SearchByKeywordResponse{
			SearchKeywords: args.Keywords,
			TotalMatches:   len(results),
			CFPs:           results,
		}, nil
	})

	// Tool 3: Find closing CFPs
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_closing_cfps",
		Description: "Find CFPs that are closing soon within a specified number of days, optionally filtered by keywords",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindClosingCFPsArgs) (*mcp.CallToolResult, FindClosingCFPsResponse, error) {
		daysAhead := args.DaysAhead
		if daysAhead == 0 {
			daysAhead = 7
		}

		openCFPs, err := getOpenCFPs()
		if err != nil {
			return nil, FindClosingCFPsResponse{}, err
		}

		now := time.Now().UnixMilli()
		maxDate := now + int64(daysAhead)*86400000

		// Filter by closing date
		var filtered []CFP
		for _, cfp := range openCFPs {
			if cfp.UntilDate <= maxDate && cfp.UntilDate >= now {
				if matchesKeywords(cfp, args.Keywords) {
					filtered = append(filtered, cfp)
				}
			}
		}

		// Sort by deadline
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].UntilDate < filtered[j].UntilDate
		})

		results := make([]CFPResult, len(filtered))
		for i, cfp := range filtered {
			results[i] = toCFPResult(cfp)
		}

		filterKeywords := "none"
		if len(args.Keywords) > 0 {
			filterKeywords = strings.Join(args.Keywords, ", ")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d urgent CFPs closing within %d days", len(results), daysAhead)},
			},
		}, FindClosingCFPsResponse{
			DaysAhead:      daysAhead,
			FilterKeywords: filterKeywords,
			UrgentCFPs:     len(results),
			CFPs:           results,
		}, nil
	})

	// Tool 4: Search by location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_cfps_by_location",
		Description: "Search for open CFPs by location/country",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchByLocationArgs) (*mcp.CallToolResult, SearchByLocationResponse, error) {
		limit := args.Limit
		if limit == 0 {
			limit = 20
		}

		openCFPs, err := getOpenCFPs()
		if err != nil {
			return nil, SearchByLocationResponse{}, err
		}

		// Filter by location
		var filtered []CFP
		for _, cfp := range openCFPs {
			if matchesLocation(cfp, args.Location) {
				filtered = append(filtered, cfp)
			}
		}

		// Sort by deadline
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].UntilDate < filtered[j].UntilDate
		})

		// Limit results
		if len(filtered) > limit {
			filtered = filtered[:limit]
		}

		results := make([]CFPResult, len(filtered))
		for i, cfp := range filtered {
			results[i] = toCFPResult(cfp)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d CFPs in %s", len(results), args.Location)},
			},
		}, SearchByLocationResponse{
			SearchLocation: args.Location,
			TotalMatches:   len(results),
			CFPs:           results,
		}, nil
	})

	return server
}

func main() {
	// Check if running in HTTP mode (Cloud Run sets PORT env var)
	port := os.Getenv("PORT")
	mode := os.Getenv("MODE")

	if port != "" || mode == "http" {
		// HTTP mode for Cloud Run
		if port == "" {
			port = "8080"
		}

		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return createServer()
		}, &mcp.StreamableHTTPOptions{
			JSONResponse: true,
		})

		// Add health check endpoint for Cloud Run
		mux := http.NewServeMux()
		mux.Handle("/", handler)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		addr := ":" + port
		log.Printf("Starting HTTP MCP server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	} else {
		// Stdio mode for local Claude Desktop
		server := createServer()
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error running server: %v\n", err)
			os.Exit(1)
		}
	}
}
