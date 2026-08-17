package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("FITNESS_WORKERS", "")
	t.Setenv("FITNESS_RETRY_LIMIT", "")
	t.Setenv("FITNESS_EXPIRE_MINUTES", "")
	t.Setenv("FITNESS_POLL_INTERVAL_MS", "")
	c := Load()
	if c.Workers != 4 {
		t.Fatalf("Workers=%d want 4", c.Workers)
	}
	if c.RetryLimit != 3 {
		t.Fatalf("RetryLimit=%d want 3", c.RetryLimit)
	}
	if len(c.SkillRoutes) == 0 {
		t.Fatal("SkillRoutes should be populated")
	}
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("FITNESS_WORKERS", "8")
	t.Setenv("FITNESS_RETRY_LIMIT", "6")
	c := Load()
	if c.Workers != 8 {
		t.Fatalf("Workers=%d want 8", c.Workers)
	}
	if c.RetryLimit != 6 {
		t.Fatalf("RetryLimit=%d want 6", c.RetryLimit)
	}
}

func TestSkillRouteFallback(t *testing.T) {
	t.Setenv("FITNESS_SKILL_ROUTES", "invalid")
	c := Load()
	if len(c.SkillRoutes) == 0 {
		t.Fatal("malformed routes should fall back to defaults")
	}
}
