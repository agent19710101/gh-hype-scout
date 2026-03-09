package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/rank"
	"github.com/agent19710101/gh-hype-scout/internal/snapshot"
)

type RankMove struct {
	FullName   string
	FromRank   int
	ToRank     int
	DeltaRank  int
	DeltaStars int
}

type DeltaReport struct {
	NewRepos []rank.Repo
	Moves    []RankMove
}

func PrintJSON(w io.Writer, in []rank.Repo) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(in)
}

func PrintTable(w io.Writer, in []rank.Repo, descWidth int) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RANK\tREPO\tSTARS\tAGE(d)\tSTARS/DAY\tSCORE\tCATEGORY\tLANG\tDESC")
	for i, r := range in {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%.1f\t%.1f\t%.1f\t%s\t%s\t%s\n", i+1, r.FullName, r.StargazersCount, r.AgeDays, r.StarsPerDay, r.HotScore, r.Category, r.Language, truncate(r.Description, descWidth))
	}
	tw.Flush()
}

func PrintThemeSummary(w io.Writer, in []rank.Repo) {
	type stat struct {
		Count int
		Sum   float64
	}
	m := map[string]stat{}
	for _, r := range in {
		s := m[r.Category]
		s.Count++
		s.Sum += r.HotScore
		m[r.Category] = s
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "THEME\tCOUNT\tAVG_SCORE")
	for _, k := range keys {
		s := m[k]
		fmt.Fprintf(tw, "%s\t%d\t%.1f\n", k, s.Count, s.Sum/float64(s.Count))
	}
	tw.Flush()
}

func BuildDelta(prev []snapshot.Item, current []rank.Repo) DeltaReport {
	report := DeltaReport{}
	if len(prev) == 0 || len(current) == 0 {
		return report
	}
	prevByName := map[string]snapshot.Item{}
	for _, p := range prev {
		prevByName[p.FullName] = p
	}
	for i, c := range current {
		old, ok := prevByName[c.FullName]
		if !ok {
			report.NewRepos = append(report.NewRepos, c)
			continue
		}
		newRank := i + 1
		if old.Rank != newRank {
			report.Moves = append(report.Moves, RankMove{FullName: c.FullName, FromRank: old.Rank, ToRank: newRank, DeltaRank: old.Rank - newRank, DeltaStars: c.StargazersCount - old.Stars})
		}
	}
	sort.Slice(report.Moves, func(i, j int) bool {
		if report.Moves[i].DeltaRank == report.Moves[j].DeltaRank {
			return report.Moves[i].ToRank < report.Moves[j].ToRank
		}
		return report.Moves[i].DeltaRank > report.Moves[j].DeltaRank
	})
	return report
}

func PrintDelta(w io.Writer, report DeltaReport) {
	if len(report.NewRepos) == 0 && len(report.Moves) == 0 {
		fmt.Fprintln(w, "\nΔ No rank changes since previous scan.")
		return
	}
	fmt.Fprintln(w, "\nΔ Changes since previous scan:")
	if len(report.NewRepos) > 0 {
		fmt.Fprintln(w, "  New repos:")
		for _, r := range report.NewRepos {
			fmt.Fprintf(w, "  + %s (%d★)\n", r.FullName, r.StargazersCount)
		}
	}
	if len(report.Moves) > 0 {
		fmt.Fprintln(w, "  Rank movers:")
		for _, m := range report.Moves {
			fmt.Fprintf(w, "  • %s %d→%d (%+d rank, %+d★)\n", m.FullName, m.FromRank, m.ToRank, m.DeltaRank, m.DeltaStars)
		}
	}
}

type move struct {
	Repo       string `json:"repo"`
	FromRank   int    `json:"from_rank"`
	ToRank     int    `json:"to_rank"`
	DeltaRank  int    `json:"delta_rank"`
	DeltaStars int    `json:"delta_stars"`
}

type deltaEvent struct {
	CapturedAt string   `json:"captured_at"`
	NewRepos   []string `json:"new_repos"`
	RankMoves  []move   `json:"rank_moves"`
}

func toDeltaEvent(now time.Time, report DeltaReport) deltaEvent {
	e := deltaEvent{CapturedAt: now.Format(time.RFC3339)}
	for _, r := range report.NewRepos {
		e.NewRepos = append(e.NewRepos, r.FullName)
	}
	for _, m := range report.Moves {
		e.RankMoves = append(e.RankMoves, move{Repo: m.FullName, FromRank: m.FromRank, ToRank: m.ToRank, DeltaRank: m.DeltaRank, DeltaStars: m.DeltaStars})
	}
	return e
}

func AppendDeltaJSONL(path string, now time.Time, report DeltaReport) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	b, err := json.Marshal(toDeltaEvent(now, report))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}

func SendDeltaWebhook(webhookURL string, now time.Time, report DeltaReport) error {
	client := &http.Client{Timeout: 4 * time.Second}
	return SendDeltaWebhookWithClient(client, webhookURL, now, report)
}

func SendDeltaWebhookWithClient(client *http.Client, webhookURL string, now time.Time, report DeltaReport) error {
	if strings.TrimSpace(webhookURL) == "" {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}
	payload, err := json.Marshal(toDeltaEvent(now, report))
	if err != nil {
		return err
	}
	var lastErr error
	delay := 250 * time.Millisecond
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return nil
		}
		if err == nil && resp != nil {
			lastErr = fmt.Errorf("webhook status %d", resp.StatusCode)
			resp.Body.Close()
		} else {
			lastErr = err
		}
		if attempt < 2 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return lastErr
}

func truncate(s string, max int) string {
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	if s == "" {
		return "-"
	}
	if max < 2 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
