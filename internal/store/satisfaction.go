package store

import (
	"context"
	"fmt"
	"sort"
)

type SatisfactionStatsFilters struct {
	StartDate string
	EndDate   string
	AvatarID  string
	Channel   string
}

type SatisfactionStats struct {
	Range                       SatisfactionStatsRange `json:"range"`
	TotalCompletedConversations int                    `json:"total_completed_conversations"`
	ReviewedConversations       int                    `json:"reviewed_conversations"`
	UnratedConversations        int                    `json:"unrated_conversations"`
	ReviewCompletionRate        float64                `json:"review_completion_rate"`
	AverageScore                float64                `json:"average_score"`
	Distribution                map[string]int         `json:"distribution"`
	ByAvatar                    []SatisfactionBucket   `json:"by_avatar"`
	ByChannel                   []SatisfactionBucket   `json:"by_channel"`
}

type SatisfactionStatsRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SatisfactionBucket struct {
	ID           string  `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Channel      string  `json:"channel,omitempty"`
	Completed    int     `json:"completed"`
	Reviewed     int     `json:"reviewed"`
	AverageScore float64 `json:"average_score"`
}

type satisfactionRow struct {
	AvatarID   string
	AvatarName string
	Channel    string
	Reviewed   bool
	Score      int
}

func (s *Store) GetSatisfactionStatistics(ctx context.Context, tenantID string, filters SatisfactionStatsFilters) (SatisfactionStats, error) {
	if s == nil || s.pg == nil {
		return SatisfactionStats{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT r.avatar_id, COALESCE(a.name,''), r.channel,
       (cr.id IS NOT NULL), COALESCE(cr.score,0)
FROM %s.conversation_records r
LEFT JOIN %s.conversation_ratings cr
  ON cr.tenant_id=r.tenant_id AND cr.call_id=r.call_id
LEFT JOIN %s.ai_avatars a ON a.id=r.avatar_id
WHERE r.tenant_id=$1
  AND r.status='archived'
  AND r.started_at >= $2::date
  AND r.started_at < ($3::date + interval '1 day')
  AND ($4='' OR r.avatar_id=$4)
  AND ($5='' OR r.channel=$5)`, schema, schema, schema),
		tenantID, filters.StartDate, filters.EndDate, filters.AvatarID, filters.Channel)
	if err != nil {
		return SatisfactionStats{}, err
	}
	defer rows.Close()
	items := make([]satisfactionRow, 0)
	for rows.Next() {
		var item satisfactionRow
		if err := rows.Scan(&item.AvatarID, &item.AvatarName, &item.Channel, &item.Reviewed, &item.Score); err != nil {
			return SatisfactionStats{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SatisfactionStats{}, err
	}
	return aggregateSatisfactionStatistics(filters, items), nil
}

func aggregateSatisfactionStatistics(filters SatisfactionStatsFilters, rows []satisfactionRow) SatisfactionStats {
	stats := SatisfactionStats{
		Range:        SatisfactionStatsRange{StartDate: filters.StartDate, EndDate: filters.EndDate},
		Distribution: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
		ByAvatar:     []SatisfactionBucket{},
		ByChannel:    []SatisfactionBucket{},
	}
	avatarBuckets := map[string]*SatisfactionBucket{}
	channelBuckets := map[string]*SatisfactionBucket{}
	for _, row := range rows {
		stats.TotalCompletedConversations++
		if row.Reviewed && row.Score >= 1 && row.Score <= 5 {
			stats.ReviewedConversations++
			stats.Distribution[fmt.Sprintf("%d", row.Score)]++
			stats.AverageScore += float64(row.Score)
		}
		avatar := avatarBuckets[row.AvatarID]
		if avatar == nil {
			avatar = &SatisfactionBucket{ID: row.AvatarID, Name: row.AvatarName}
			avatarBuckets[row.AvatarID] = avatar
		}
		avatar.Completed++
		if row.Reviewed && row.Score >= 1 && row.Score <= 5 {
			avatar.Reviewed++
			avatar.AverageScore += float64(row.Score)
		}
		channel := channelBuckets[row.Channel]
		if channel == nil {
			channel = &SatisfactionBucket{Channel: row.Channel}
			channelBuckets[row.Channel] = channel
		}
		channel.Completed++
		if row.Reviewed && row.Score >= 1 && row.Score <= 5 {
			channel.Reviewed++
			channel.AverageScore += float64(row.Score)
		}
	}
	stats.UnratedConversations = stats.TotalCompletedConversations - stats.ReviewedConversations
	if stats.TotalCompletedConversations > 0 {
		stats.ReviewCompletionRate = roundSatisfaction(float64(stats.ReviewedConversations) * 100 / float64(stats.TotalCompletedConversations))
	}
	if stats.ReviewedConversations > 0 {
		stats.AverageScore = roundSatisfaction(stats.AverageScore / float64(stats.ReviewedConversations))
	}
	for _, bucket := range avatarBuckets {
		if bucket.Reviewed > 0 {
			bucket.AverageScore = roundSatisfaction(bucket.AverageScore / float64(bucket.Reviewed))
		}
		stats.ByAvatar = append(stats.ByAvatar, *bucket)
	}
	for _, bucket := range channelBuckets {
		if bucket.Reviewed > 0 {
			bucket.AverageScore = roundSatisfaction(bucket.AverageScore / float64(bucket.Reviewed))
		}
		stats.ByChannel = append(stats.ByChannel, *bucket)
	}
	sort.Slice(stats.ByAvatar, func(i, j int) bool { return stats.ByAvatar[i].Name < stats.ByAvatar[j].Name })
	sort.Slice(stats.ByChannel, func(i, j int) bool { return stats.ByChannel[i].Channel < stats.ByChannel[j].Channel })
	return stats
}

func roundSatisfaction(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
