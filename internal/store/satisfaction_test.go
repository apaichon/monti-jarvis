package store

import "testing"

func TestAggregateSatisfactionStatistics(t *testing.T) {
	stats := aggregateSatisfactionStatistics(SatisfactionStatsFilters{
		StartDate: "2026-07-14",
		EndDate:   "2026-07-14",
	}, []satisfactionRow{
		{AvatarID: "ava", AvatarName: "Ava", Channel: "voice", Reviewed: true, Score: 5},
		{AvatarID: "ava", AvatarName: "Ava", Channel: "voice", Reviewed: true, Score: 4},
		{AvatarID: "neo", AvatarName: "Neo", Channel: "chat", Reviewed: false},
	})

	if stats.TotalCompletedConversations != 3 || stats.ReviewedConversations != 2 || stats.UnratedConversations != 1 {
		t.Fatalf("unexpected counts: %+v", stats)
	}
	if stats.ReviewCompletionRate != 66.67 || stats.AverageScore != 4.5 {
		t.Fatalf("unexpected averages: %+v", stats)
	}
	if stats.Distribution["5"] != 1 || stats.Distribution["4"] != 1 || stats.Distribution["1"] != 0 {
		t.Fatalf("unexpected distribution: %+v", stats.Distribution)
	}
	if len(stats.ByAvatar) != 2 || stats.ByAvatar[0].Name != "Ava" || stats.ByAvatar[0].AverageScore != 4.5 {
		t.Fatalf("unexpected avatar buckets: %+v", stats.ByAvatar)
	}
	if len(stats.ByChannel) != 2 || stats.ByChannel[0].Channel != "chat" || stats.ByChannel[1].Reviewed != 2 {
		t.Fatalf("unexpected channel buckets: %+v", stats.ByChannel)
	}
}

func TestAggregateSatisfactionStatisticsEmpty(t *testing.T) {
	stats := aggregateSatisfactionStatistics(SatisfactionStatsFilters{StartDate: "2026-07-14", EndDate: "2026-07-14"}, nil)
	if stats.TotalCompletedConversations != 0 || stats.ReviewCompletionRate != 0 || stats.AverageScore != 0 {
		t.Fatalf("unexpected empty stats: %+v", stats)
	}
	if len(stats.ByAvatar) != 0 || len(stats.ByChannel) != 0 || len(stats.Distribution) != 5 {
		t.Fatalf("unexpected empty buckets: %+v", stats)
	}
}
