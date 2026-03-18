package service

import (
	"testing"

	"swd-new/internal/repository"

	"swd-new/pkg/test"
)

var TestSensitiveWordService SensitiveWordService

func TestMain(m *testing.M) {
	env, err := test.SetupTestEnvironment()
	if err != nil {
		panic(err)
	}
	testDB, err := repository.NewSensitiveWordRepository(env.TestDB)
	if err != nil {
		panic(err)
	}
	testService := NewService(env.TestLogger)
	TestSensitiveWordService, err = NewSensitiveWordService(testService, testDB)
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestSensitiveWordMatchAndReplace(t *testing.T) {
	resp, err := TestSensitiveWordService.Check("你这个蠢猪")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !resp.Contains {
		t.Fatalf("expected contains to be true")
	}

	if resp.FilteredText != "你这个**" {
		t.Fatalf("unexpected filtered text: %s", resp.FilteredText)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("unexpected match count: %d", len(resp.Matches))
	}

	if resp.Matches[0].Word != "蠢猪" || resp.Matches[0].StartPos != 3 || resp.Matches[0].EndPos != 5 {
		t.Fatalf("unexpected first match: %+v", resp.Matches[0])
	}
}

func BenchmarkCheckSensitiveWords(b *testing.B) {
	for b.Loop() {
		TestSensitiveWordService.Check("你这个蠢猪真是个坏蛋")
	}
}
