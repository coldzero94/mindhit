# Phase 9: AI 마인드맵 생성

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | OpenAI API를 사용한 콘텐츠 요약 및 마인드맵 생성 |
| **선행 조건** | Phase 6 완료 (스케줄러) |
| **예상 소요** | 4 Steps |
| **결과물** | 세션 완료 시 자동 마인드맵 생성 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 9.1 | OpenAI 클라이언트 설정 | ⬜ |
| 9.2 | URL 요약 서비스 | ⬜ |
| 9.3 | 마인드맵 생성 서비스 | ⬜ |
| 9.4 | AI 처리 Job 등록 | ⬜ |

---

## Step 9.1: OpenAI 클라이언트 설정

### 체크리스트

- [ ] **의존성 추가**
  ```bash
  go get github.com/sashabaranov/go-openai
  ```

- [ ] **환경 변수 추가**
  ```env
  OPENAI_API_KEY=sk-...
  OPENAI_MODEL=gpt-4-turbo-preview
  ```

- [ ] **Config 업데이트**
  - [ ] `internal/infrastructure/config/config.go`
    ```go
    type Config struct {
        // ... 기존 필드

        // OpenAI
        OpenAIAPIKey string
        OpenAIModel  string
    }

    func Load() *Config {
        return &Config{
            // ... 기존 필드

            OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
            OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4-turbo-preview"),
        }
    }
    ```

- [ ] **OpenAI 클라이언트 래퍼**
  - [ ] `internal/infrastructure/ai/openai.go`
    ```go
    package ai

    import (
        "context"
        "fmt"

        "github.com/sashabaranov/go-openai"
    )

    type OpenAIClient struct {
        client *openai.Client
        model  string
    }

    func NewOpenAIClient(apiKey, model string) *OpenAIClient {
        return &OpenAIClient{
            client: openai.NewClient(apiKey),
            model:  model,
        }
    }

    type Message struct {
        Role    string
        Content string
    }

    func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
        chatMessages := make([]openai.ChatCompletionMessage, len(messages))
        for i, msg := range messages {
            chatMessages[i] = openai.ChatCompletionMessage{
                Role:    msg.Role,
                Content: msg.Content,
            }
        }

        resp, err := c.client.CreateChatCompletion(
            ctx,
            openai.ChatCompletionRequest{
                Model:    c.model,
                Messages: chatMessages,
            },
        )
        if err != nil {
            return "", fmt.Errorf("openai chat completion: %w", err)
        }

        if len(resp.Choices) == 0 {
            return "", fmt.Errorf("no response from openai")
        }

        return resp.Choices[0].Message.Content, nil
    }

    func (c *OpenAIClient) ChatWithJSON(ctx context.Context, messages []Message) (string, error) {
        chatMessages := make([]openai.ChatCompletionMessage, len(messages))
        for i, msg := range messages {
            chatMessages[i] = openai.ChatCompletionMessage{
                Role:    msg.Role,
                Content: msg.Content,
            }
        }

        resp, err := c.client.CreateChatCompletion(
            ctx,
            openai.ChatCompletionRequest{
                Model:    c.model,
                Messages: chatMessages,
                ResponseFormat: &openai.ChatCompletionResponseFormat{
                    Type: openai.ChatCompletionResponseFormatTypeJSONObject,
                },
            },
        )
        if err != nil {
            return "", fmt.Errorf("openai chat completion: %w", err)
        }

        if len(resp.Choices) == 0 {
            return "", fmt.Errorf("no response from openai")
        }

        return resp.Choices[0].Message.Content, nil
    }
    ```

- [ ] **main.go에 OpenAI 클라이언트 추가**
  ```go
  import "github.com/mindhit/api/internal/infrastructure/ai"

  // Initialize OpenAI client
  openaiClient := ai.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
  ```

### 검증
```bash
go build ./...
# 컴파일 성공
```

---

## Step 9.2: URL 요약 서비스

### 체크리스트

- [ ] **URL 콘텐츠 가져오기**
  - [ ] `internal/infrastructure/crawler/crawler.go`
    ```go
    package crawler

    import (
        "context"
        "fmt"
        "io"
        "net/http"
        "strings"
        "time"

        "github.com/PuerkitoBio/goquery"
    )

    type Crawler struct {
        client *http.Client
    }

    func New() *Crawler {
        return &Crawler{
            client: &http.Client{
                Timeout: 30 * time.Second,
            },
        }
    }

    type PageContent struct {
        Title   string
        Content string
        URL     string
    }

    func (c *Crawler) FetchContent(ctx context.Context, url string) (*PageContent, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }

        req.Header.Set("User-Agent", "MindHit/1.0 (Content Summarizer)")

        resp, err := c.client.Do(req)
        if err != nil {
            return nil, fmt.Errorf("fetch url: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
        }

        body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
        if err != nil {
            return nil, fmt.Errorf("read body: %w", err)
        }

        doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
        if err != nil {
            return nil, fmt.Errorf("parse html: %w", err)
        }

        // Remove script, style, nav, footer elements
        doc.Find("script, style, nav, footer, header, aside").Remove()

        title := doc.Find("title").First().Text()
        if title == "" {
            title = doc.Find("h1").First().Text()
        }

        // Extract main content
        var content string
        mainSelectors := []string{
            "article",
            "main",
            "[role='main']",
            ".content",
            ".post-content",
            ".article-content",
        }

        for _, selector := range mainSelectors {
            if sel := doc.Find(selector).First(); sel.Length() > 0 {
                content = sel.Text()
                break
            }
        }

        if content == "" {
            content = doc.Find("body").Text()
        }

        // Clean up content
        content = strings.TrimSpace(content)
        content = strings.Join(strings.Fields(content), " ")

        // Limit content length
        if len(content) > 10000 {
            content = content[:10000] + "..."
        }

        return &PageContent{
            Title:   strings.TrimSpace(title),
            Content: content,
            URL:     url,
        }, nil
    }
    ```

- [ ] **의존성 추가**
  ```bash
  go get github.com/PuerkitoBio/goquery
  ```

- [ ] **URL 요약 서비스**
  - [ ] `internal/service/summarize_service.go`
    ```go
    package service

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/url"
        "github.com/mindhit/api/internal/infrastructure/ai"
        "github.com/mindhit/api/internal/infrastructure/crawler"
    )

    type SummarizeService struct {
        client   *ent.Client
        ai       *ai.OpenAIClient
        crawler  *crawler.Crawler
    }

    func NewSummarizeService(client *ent.Client, ai *ai.OpenAIClient, crawler *crawler.Crawler) *SummarizeService {
        return &SummarizeService{
            client:  client,
            ai:      ai,
            crawler: crawler,
        }
    }

    type URLSummary struct {
        Summary  string   `json:"summary"`
        Keywords []string `json:"keywords"`
    }

    const summarizePrompt = `You are a content summarizer. Analyze the following web page content and provide:
1. A concise summary (2-3 sentences) in Korean
2. 3-5 relevant keywords in Korean

Respond in JSON format:
{
  "summary": "요약 내용",
  "keywords": ["키워드1", "키워드2", "키워드3"]
}

Web page title: %s
Web page content:
%s`

    func (s *SummarizeService) SummarizeURL(ctx context.Context, urlID uuid.UUID) error {
        // Get URL from database
        u, err := s.client.URL.Get(ctx, urlID)
        if err != nil {
            return fmt.Errorf("get url: %w", err)
        }

        // Skip if already summarized
        if u.Summary != "" {
            return nil
        }

        // Check if content needs to be re-crawled (older than 24 hours)
        needsCrawl := u.Content == "" || u.CrawledAt == nil ||
            time.Since(*u.CrawledAt) > 24*time.Hour

        var content *crawler.PageContent
        if needsCrawl {
            // Fetch content
            var err error
            content, err = s.crawler.FetchContent(ctx, u.URL)
            if err != nil {
                slog.Warn("failed to fetch url content", "url", u.URL, "error", err)
                return nil // Don't fail the whole process
            }
        } else {
            // Use existing content
            content = &crawler.PageContent{
                Title:   u.Title,
                Content: u.Content,
                URL:     u.URL,
            }
        }

        // Generate summary using AI
        messages := []ai.Message{
            {
                Role:    "user",
                Content: fmt.Sprintf(summarizePrompt, content.Title, content.Content),
            },
        }

        response, err := s.ai.ChatWithJSON(ctx, messages)
        if err != nil {
            return fmt.Errorf("ai summarize: %w", err)
        }

        var summary URLSummary
        if err := json.Unmarshal([]byte(response), &summary); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Update URL in database
        now := time.Now()
        update := s.client.URL.UpdateOneID(urlID).
            SetSummary(summary.Summary).
            SetKeywords(summary.Keywords)

        // Only update content and crawled_at if we actually crawled
        if needsCrawl {
            update = update.
                SetTitle(content.Title).
                SetContent(content.Content).
                SetCrawledAt(now)
        }

        _, err = update.Save(ctx)
        if err != nil {
            return fmt.Errorf("update url: %w", err)
        }

        slog.Info("summarized url", "url", u.URL, "keywords", summary.Keywords)
        return nil
    }

    func (s *SummarizeService) SummarizeSessionURLs(ctx context.Context, sessionID uuid.UUID) error {
        // Get all page visits for session
        pageVisits, err := s.client.PageVisit.
            Query().
            Where(/* session_id = sessionID */).
            WithURL().
            All(ctx)

        if err != nil {
            return fmt.Errorf("get page visits: %w", err)
        }

        for _, pv := range pageVisits {
            if pv.Edges.URL == nil {
                continue
            }

            if err := s.SummarizeURL(ctx, pv.Edges.URL.ID); err != nil {
                slog.Error("failed to summarize url",
                    "url_id", pv.Edges.URL.ID,
                    "error", err,
                )
                // Continue with other URLs
            }
        }

        return nil
    }
    ```

### 검증
```bash
go build ./...
# 컴파일 성공

# 테스트 (유닛 테스트)
go test ./internal/service/...
```

---

## Step 9.3: 마인드맵 생성 서비스

### 체크리스트

- [ ] **마인드맵 타입 정의**
  - [ ] `internal/service/mindmap_types.go`
    ```go
    package service

    type MindmapNode struct {
        ID       string                 `json:"id"`
        Label    string                 `json:"label"`
        Type     string                 `json:"type"` // core, topic, subtopic, page
        Size     float64                `json:"size"`
        Color    string                 `json:"color"`
        Position *Position              `json:"position,omitempty"`
        Data     map[string]interface{} `json:"data"`
    }

    type Position struct {
        X float64 `json:"x"`
        Y float64 `json:"y"`
        Z float64 `json:"z"`
    }

    type MindmapEdge struct {
        Source string  `json:"source"`
        Target string  `json:"target"`
        Weight float64 `json:"weight"`
    }

    type MindmapLayout struct {
        Type   string `json:"type"` // galaxy, tree, radial
        Params map[string]interface{} `json:"params"`
    }

    type MindmapData struct {
        Nodes  []MindmapNode `json:"nodes"`
        Edges  []MindmapEdge `json:"edges"`
        Layout MindmapLayout `json:"layout"`
    }
    ```

- [ ] **마인드맵 생성 서비스**
  - [ ] `internal/service/mindmap_service.go`
    ```go
    package service

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"
        "math"
        "strings"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/ai"
    )

    type MindmapService struct {
        client *ent.Client
        ai     *ai.OpenAIClient
    }

    func NewMindmapService(client *ent.Client, ai *ai.OpenAIClient) *MindmapService {
        return &MindmapService{
            client: client,
            ai:     ai,
        }
    }

    const mindmapPrompt = `Analyze the following browsing session data and generate a mindmap structure.

Session contains the following URLs with their summaries:
%s

Highlights from the session:
%s

Generate a hierarchical mindmap structure with:
1. One core topic (central theme of the browsing session)
2. 3-5 main topics (major themes)
3. Sub-topics under each main topic (related pages/concepts)

Respond in JSON format:
{
  "core": {
    "label": "핵심 주제",
    "description": "설명"
  },
  "topics": [
    {
      "label": "주요 토픽",
      "description": "설명",
      "subtopics": [
        {"label": "하위 토픽", "url_ids": ["uuid1", "uuid2"]}
      ]
    }
  ],
  "connections": [
    {"from": "토픽1", "to": "토픽2", "reason": "연결 이유"}
  ]
}`

    type AIResponse struct {
        Core struct {
            Label       string `json:"label"`
            Description string `json:"description"`
        } `json:"core"`
        Topics []struct {
            Label       string `json:"label"`
            Description string `json:"description"`
            Subtopics   []struct {
                Label  string   `json:"label"`
                URLIDs []string `json:"url_ids"`
            } `json:"subtopics"`
        } `json:"topics"`
        Connections []struct {
            From   string `json:"from"`
            To     string `json:"to"`
            Reason string `json:"reason"`
        } `json:"connections"`
    }

    func (s *MindmapService) GenerateMindmap(ctx context.Context, sessionID uuid.UUID) error {
        // Get session with all related data
        sess, err := s.client.Session.
            Query().
            Where(session.IDEQ(sessionID)).
            WithPageVisits(func(q *ent.PageVisitQuery) {
                q.WithURL()
            }).
            WithHighlights().
            Only(ctx)

        if err != nil {
            return fmt.Errorf("get session: %w", err)
        }

        // Build URL summaries text
        var urlSummaries strings.Builder
        urlMap := make(map[string]*ent.URL)
        dwellTimeMap := make(map[string]int)

        for _, pv := range sess.Edges.PageVisits {
            if pv.Edges.URL == nil {
                continue
            }
            u := pv.Edges.URL
            urlMap[u.ID.String()] = u
            dwellTimeMap[u.ID.String()] = pv.DwellTimeSeconds

            urlSummaries.WriteString(fmt.Sprintf("- [%s] %s\n  URL: %s\n  Summary: %s\n  Keywords: %s\n  Dwell time: %d seconds\n\n",
                u.ID.String(),
                u.Title,
                u.URL,
                u.Summary,
                strings.Join(u.Keywords, ", "),
                pv.DwellTimeSeconds,
            ))
        }

        // Build highlights text
        var highlights strings.Builder
        for _, h := range sess.Edges.Highlights {
            highlights.WriteString(fmt.Sprintf("- \"%s\"\n", h.Text))
        }

        // Generate mindmap structure using AI
        messages := []ai.Message{
            {
                Role:    "user",
                Content: fmt.Sprintf(mindmapPrompt, urlSummaries.String(), highlights.String()),
            },
        }

        response, err := s.ai.ChatWithJSON(ctx, messages)
        if err != nil {
            return fmt.Errorf("ai generate mindmap: %w", err)
        }

        var aiResp AIResponse
        if err := json.Unmarshal([]byte(response), &aiResp); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Convert AI response to mindmap data
        mindmapData := s.buildMindmapData(aiResp, urlMap, dwellTimeMap)

        // Save mindmap to database
        _, err = s.client.MindmapGraph.
            Create().
            SetSessionID(sessionID).
            SetNodes(mindmapData.Nodes).
            SetEdges(mindmapData.Edges).
            SetLayout(mindmapData.Layout).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("save mindmap: %w", err)
        }

        // Update session status
        _, err = s.client.Session.
            UpdateOneID(sessionID).
            SetStatus(session.StatusCompleted).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("update session status: %w", err)
        }

        slog.Info("generated mindmap", "session_id", sessionID)
        return nil
    }

    func (s *MindmapService) buildMindmapData(resp AIResponse, urlMap map[string]*ent.URL, dwellTimeMap map[string]int) MindmapData {
        var nodes []MindmapNode
        var edges []MindmapEdge
        nodeIDMap := make(map[string]string)

        // Create core node (center of galaxy)
        coreID := uuid.New().String()
        nodes = append(nodes, MindmapNode{
            ID:    coreID,
            Label: resp.Core.Label,
            Type:  "core",
            Size:  100,
            Color: "#FFD700", // Gold for core
            Position: &Position{X: 0, Y: 0, Z: 0},
            Data: map[string]interface{}{
                "description": resp.Core.Description,
            },
        })
        nodeIDMap[resp.Core.Label] = coreID

        // Create topic nodes (planets orbiting the sun)
        topicCount := len(resp.Topics)
        for i, topic := range resp.Topics {
            topicID := uuid.New().String()
            nodeIDMap[topic.Label] = topicID

            // Position in orbit around core
            angle := (float64(i) / float64(topicCount)) * 2 * math.Pi
            radius := 200.0

            nodes = append(nodes, MindmapNode{
                ID:    topicID,
                Label: topic.Label,
                Type:  "topic",
                Size:  60,
                Color: getTopicColor(i),
                Position: &Position{
                    X: radius * math.Cos(angle),
                    Y: radius * math.Sin(angle),
                    Z: 0,
                },
                Data: map[string]interface{}{
                    "description": topic.Description,
                },
            })

            // Edge from core to topic
            edges = append(edges, MindmapEdge{
                Source: coreID,
                Target: topicID,
                Weight: 1.0,
            })

            // Create subtopic nodes (moons)
            for j, subtopic := range topic.Subtopics {
                subtopicID := uuid.New().String()

                // Position around parent topic
                subAngle := angle + (float64(j)-float64(len(topic.Subtopics))/2)*0.3
                subRadius := 80.0

                // Calculate size based on dwell time
                size := 20.0
                for _, urlID := range subtopic.URLIDs {
                    if dwell, ok := dwellTimeMap[urlID]; ok {
                        size = math.Min(50, 20+float64(dwell)/10)
                    }
                }

                nodes = append(nodes, MindmapNode{
                    ID:    subtopicID,
                    Label: subtopic.Label,
                    Type:  "subtopic",
                    Size:  size,
                    Color: getTopicColor(i),
                    Position: &Position{
                        X: radius*math.Cos(angle) + subRadius*math.Cos(subAngle),
                        Y: radius*math.Sin(angle) + subRadius*math.Sin(subAngle),
                        Z: 0,
                    },
                    Data: map[string]interface{}{
                        "url_ids": subtopic.URLIDs,
                    },
                })

                edges = append(edges, MindmapEdge{
                    Source: topicID,
                    Target: subtopicID,
                    Weight: 0.5,
                })
            }
        }

        // Add cross-topic connections
        for _, conn := range resp.Connections {
            fromID, fromOK := nodeIDMap[conn.From]
            toID, toOK := nodeIDMap[conn.To]
            if fromOK && toOK {
                edges = append(edges, MindmapEdge{
                    Source: fromID,
                    Target: toID,
                    Weight: 0.3,
                })
            }
        }

        return MindmapData{
            Nodes: nodes,
            Edges: edges,
            Layout: MindmapLayout{
                Type: "galaxy",
                Params: map[string]interface{}{
                    "center": []float64{0, 0, 0},
                    "scale":  1.0,
                },
            },
        }
    }

    func getTopicColor(index int) string {
        colors := []string{
            "#3B82F6", // Blue
            "#10B981", // Green
            "#F59E0B", // Amber
            "#EF4444", // Red
            "#8B5CF6", // Purple
            "#EC4899", // Pink
            "#14B8A6", // Teal
            "#F97316", // Orange
        }
        return colors[index%len(colors)]
    }
    ```

### 검증
```bash
go build ./...
# 컴파일 성공
```

---

## Step 9.4: AI 처리 Job 등록

### 체크리스트

- [ ] **AI 처리 Job**
  - [ ] `internal/jobs/ai_processing.go`
    ```go
    package jobs

    import (
        "context"
        "log/slog"
        "time"

        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/service"
    )

    type AIProcessingJob struct {
        client            *ent.Client
        summarizeService  *service.SummarizeService
        mindmapService    *service.MindmapService
    }

    func NewAIProcessingJob(
        client *ent.Client,
        summarizeService *service.SummarizeService,
        mindmapService *service.MindmapService,
    ) *AIProcessingJob {
        return &AIProcessingJob{
            client:           client,
            summarizeService: summarizeService,
            mindmapService:   mindmapService,
        }
    }

    func (j *AIProcessingJob) Run() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
        defer cancel()

        // Find sessions in "processing" status
        sessions, err := j.client.Session.
            Query().
            Where(session.StatusEQ(session.StatusProcessing)).
            Limit(5). // Process 5 at a time
            All(ctx)

        if err != nil {
            slog.Error("failed to get processing sessions", "error", err)
            return
        }

        for _, sess := range sessions {
            slog.Info("processing session", "session_id", sess.ID)

            // Step 1: Summarize all URLs
            if err := j.summarizeService.SummarizeSessionURLs(ctx, sess.ID); err != nil {
                slog.Error("failed to summarize session urls",
                    "session_id", sess.ID,
                    "error", err,
                )
                j.markSessionFailed(ctx, sess.ID)
                continue
            }

            // Step 2: Generate mindmap
            if err := j.mindmapService.GenerateMindmap(ctx, sess.ID); err != nil {
                slog.Error("failed to generate mindmap",
                    "session_id", sess.ID,
                    "error", err,
                )
                j.markSessionFailed(ctx, sess.ID)
                continue
            }

            slog.Info("session processing completed", "session_id", sess.ID)
        }
    }

    func (j *AIProcessingJob) markSessionFailed(ctx context.Context, sessionID uuid.UUID) {
        _, err := j.client.Session.
            UpdateOneID(sessionID).
            SetStatus(session.StatusFailed).
            Save(ctx)

        if err != nil {
            slog.Error("failed to mark session as failed",
                "session_id", sessionID,
                "error", err,
            )
        }
    }
    ```

- [ ] **main.go에 AI Job 등록**
  ```go
  import (
      "github.com/mindhit/api/internal/infrastructure/ai"
      "github.com/mindhit/api/internal/infrastructure/crawler"
      "github.com/mindhit/api/internal/jobs"
      "github.com/mindhit/api/internal/service"
  )

  // Initialize AI components
  openaiClient := ai.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
  crawlerClient := crawler.New()

  // Initialize services
  summarizeService := service.NewSummarizeService(client, openaiClient, crawlerClient)
  mindmapService := service.NewMindmapService(client, openaiClient)

  // Register AI processing job (every 5 minutes)
  aiProcessingJob := jobs.NewAIProcessingJob(client, summarizeService, mindmapService)
  if err := sched.RegisterIntervalJob("ai-processing", 5*time.Minute, aiProcessingJob.Run); err != nil {
      slog.Error("failed to register ai processing job", "error", err)
  }
  ```

- [ ] **마인드맵 API 엔드포인트**
  - [ ] `internal/controller/mindmap_controller.go`
    ```go
    package controller

    import (
        "net/http"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/mindmapgraph"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/middleware"
        "github.com/mindhit/api/internal/service"
    )

    type MindmapController struct {
        client           *ent.Client
        summarizeService *service.SummarizeService
        mindmapService   *service.MindmapService
    }

    func NewMindmapController(
        client *ent.Client,
        summarizeService *service.SummarizeService,
        mindmapService *service.MindmapService,
    ) *MindmapController {
        return &MindmapController{
            client:           client,
            summarizeService: summarizeService,
            mindmapService:   mindmapService,
        }
    }

    // GetBySession retrieves the mindmap for a session
    func (c *MindmapController) GetBySession(ctx *gin.Context) {
        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        mindmap, err := c.client.MindmapGraph.
            Query().
            Where(mindmapgraph.SessionIDEQ(sessionID)).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "mindmap not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        ctx.JSON(http.StatusOK, gin.H{"mindmap": mindmap})
    }

    // Generate triggers mindmap generation for a session
    func (c *MindmapController) Generate(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized"}})
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        // Verify session ownership and status
        sess, err := c.client.Session.
            Query().
            Where(
                session.IDEQ(sessionID),
                session.HasUserWith(user.IDEQ(userID)),
            ).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "session not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Check if session is in valid state for mindmap generation
        if sess.Status != session.StatusProcessing && sess.Status != session.StatusCompleted {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "session must be stopped before generating mindmap"},
            })
            return
        }

        // Check if mindmap already exists
        exists, err := c.client.MindmapGraph.
            Query().
            Where(mindmapgraph.SessionIDEQ(sessionID)).
            Exist(ctx.Request.Context())

        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        if exists {
            ctx.JSON(http.StatusConflict, gin.H{
                "error": gin.H{"message": "mindmap already exists for this session"},
            })
            return
        }

        // Start async processing (or queue it)
        go func() {
            bgCtx := context.Background()

            // Summarize URLs first
            if err := c.summarizeService.SummarizeSessionURLs(bgCtx, sessionID); err != nil {
                slog.Error("failed to summarize session urls", "session_id", sessionID, "error", err)
                return
            }

            // Generate mindmap
            if err := c.mindmapService.GenerateMindmap(bgCtx, sessionID); err != nil {
                slog.Error("failed to generate mindmap", "session_id", sessionID, "error", err)
                return
            }
        }()

        ctx.JSON(http.StatusAccepted, gin.H{
            "message":    "mindmap generation started",
            "session_id": sessionID.String(),
        })
    }

    // Regenerate regenerates the mindmap for a session (replaces existing)
    func (c *MindmapController) Regenerate(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized"}})
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        // Verify session ownership
        _, err = c.client.Session.
            Query().
            Where(
                session.IDEQ(sessionID),
                session.HasUserWith(user.IDEQ(userID)),
            ).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "session not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Delete existing mindmap if exists
        _, err = c.client.MindmapGraph.
            Delete().
            Where(mindmapgraph.SessionIDEQ(sessionID)).
            Exec(ctx.Request.Context())

        if err != nil && !ent.IsNotFound(err) {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Start async processing
        go func() {
            bgCtx := context.Background()

            if err := c.mindmapService.GenerateMindmap(bgCtx, sessionID); err != nil {
                slog.Error("failed to regenerate mindmap", "session_id", sessionID, "error", err)
                return
            }
        }()

        ctx.JSON(http.StatusAccepted, gin.H{
            "message":    "mindmap regeneration started",
            "session_id": sessionID.String(),
        })
    }
    ```

- [ ] **라우터에 마인드맵 엔드포인트 추가**
  ```go
  // In main.go or router setup
  mindmapController := controller.NewMindmapController(client, summarizeService, mindmapService)

  sessions := v1.Group("/sessions")
  sessions.Use(middleware.Auth(jwtService))
  {
      // ... 기존 세션 라우트

      // Mindmap
      sessions.GET("/:id/mindmap", mindmapController.GetBySession)
      sessions.POST("/:id/mindmap/generate", mindmapController.Generate)
      sessions.POST("/:id/mindmap/regenerate", mindmapController.Regenerate)
  }
  ```

### 검증
```bash
# 서버 실행
go run ./cmd/server

# 세션 종료 후 5분 이내 마인드맵 생성 확인
# 로그에서 "session processing completed" 메시지 확인

# API로 마인드맵 조회
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/sessions/{session_id}/mindmap
```

---

## API 요약

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/v1/sessions/:id/mindmap` | 세션의 마인드맵 조회 | Bearer Token |
| POST | `/v1/sessions/:id/mindmap/generate` | 마인드맵 생성 시작 (비동기) | Bearer Token |
| POST | `/v1/sessions/:id/mindmap/regenerate` | 마인드맵 재생성 (기존 삭제 후 새로 생성) | Bearer Token |

### 응답 예시

**GET /v1/sessions/:id/mindmap** (성공)
```json
{
  "mindmap": {
    "id": "uuid",
    "session_id": "uuid",
    "nodes": [
      {
        "id": "node-1",
        "label": "핵심 주제",
        "type": "core",
        "size": 100,
        "color": "#FFD700",
        "position": {"x": 0, "y": 0, "z": 0}
      }
    ],
    "edges": [
      {"source": "node-1", "target": "node-2", "weight": 1.0}
    ],
    "layout": {
      "type": "galaxy",
      "params": {"center": [0, 0, 0], "scale": 1.0}
    },
    "generated_at": "2024-01-01T00:00:00Z"
  }
}
```

**POST /v1/sessions/:id/mindmap/generate** (성공 - 202 Accepted)
```json
{
  "message": "mindmap generation started",
  "session_id": "uuid"
}
```

**POST /v1/sessions/:id/mindmap/regenerate** (성공 - 202 Accepted)
```json
{
  "message": "mindmap regeneration started",
  "session_id": "uuid"
}
```

---

## Phase 9 완료 확인

### 전체 검증 체크리스트

- [ ] OpenAI API 연동
- [ ] URL 콘텐츠 크롤링
- [ ] URL 요약 생성
- [ ] 마인드맵 구조 생성
- [ ] 마인드맵 저장
- [ ] 마인드맵 API 조회
- [ ] 5분마다 자동 처리

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| OpenAI 클라이언트 | `internal/infrastructure/ai/openai.go` |
| 크롤러 | `internal/infrastructure/crawler/crawler.go` |
| 요약 서비스 | `internal/service/summarize_service.go` |
| 마인드맵 서비스 | `internal/service/mindmap_service.go` |
| AI 처리 Job | `internal/jobs/ai_processing.go` |
| 마인드맵 API | `internal/controller/mindmap_controller.go` |

---

## 다음 Phase

Phase 9 완료 후 [Phase 10: 웹앱 대시보드](./phase-10-dashboard.md)으로 진행하세요.
