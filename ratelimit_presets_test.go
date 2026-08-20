package snipeit

import "testing"

func TestRateLimitPresetValues(t *testing.T) {
	for _, tc := range []struct {
		preset  RateLimitPreset
		name    string
		perMin  int
		wantRPS float64 // 0 means no limiter
	}{
		{PresetBasic, "basic", 120, 2},
		{PresetSmallBusiness, "small_business", 240, 4},
		{PresetDedicated, "dedicated", 0, 0},
	} {
		if tc.preset.Name != tc.name || tc.preset.RequestsPerMinute != tc.perMin {
			t.Errorf("preset %+v: want name %q and %d req/min", tc.preset, tc.name, tc.perMin)
		}

		limiter := tc.preset.Limiter()
		if tc.wantRPS == 0 {
			if limiter != nil {
				t.Errorf("%s: expected nil limiter for unmetered plan, got %T", tc.name, limiter)
			}
			continue
		}
		adaptive, ok := limiter.(*AdaptiveRateLimiter)
		if !ok {
			t.Fatalf("%s: expected *AdaptiveRateLimiter, got %T", tc.name, limiter)
		}
		if _, rate := adaptive.Snapshot(); rate != tc.wantRPS {
			t.Errorf("%s: expected %v req/s, got %v", tc.name, tc.wantRPS, rate)
		}
	}
}

func TestPresetByName(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  RateLimitPreset
	}{
		{"basic", PresetBasic},
		{"BASIC", PresetBasic},
		{"small_business", PresetSmallBusiness},
		{"small-business", PresetSmallBusiness},
		{"SmallBusiness", PresetSmallBusiness},
		{" dedicated ", PresetDedicated},
	} {
		got, ok := PresetByName(tc.input)
		if !ok || got != tc.want {
			t.Errorf("PresetByName(%q) = %+v, %v; want %+v, true", tc.input, got, ok, tc.want)
		}
	}

	if got, ok := PresetByName("enterprise"); ok {
		t.Errorf("PresetByName(\"enterprise\") = %+v, true; want zero value, false", got)
	}
}
