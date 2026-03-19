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

	if !resp.Contain {
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
	t.Logf("Check result: %+v", resp)
}

func TestSensitiveRefreshesDictionaryAfterCRUD(t *testing.T) {
	word, err := TestSensitiveWordService.CreateWord(t.Context(), CreateSensitiveWordInput{
		Word: "测试脏词A",
		Type: DefaultSensitiveWordType,
	})
	if err != nil {
		t.Fatalf("CreateWord failed: %v", err)
	}

	resp, err := TestSensitiveWordService.Check("这里有测试脏词A")
	if err != nil {
		t.Fatalf("Check after create failed: %v", err)
	}
	if !resp.Contain {
		t.Fatalf("expected created word to be detected")
	}

	word, err = TestSensitiveWordService.UpdateWord(t.Context(), word.ID, UpdateSensitiveWordInput{
		Word: "测试脏词B",
		Type: DefaultSensitiveWordType,
	})
	if err != nil {
		t.Fatalf("UpdateWord failed: %v", err)
	}

	resp, err = TestSensitiveWordService.Check("这里有测试脏词A")
	if err != nil {
		t.Fatalf("Check old word after update failed: %v", err)
	}
	if resp.Contain {
		t.Fatalf("expected old word to be removed after update")
	}

	resp, err = TestSensitiveWordService.Check("这里有测试脏词B")
	if err != nil {
		t.Fatalf("Check new word after update failed: %v", err)
	}
	if !resp.Contain {
		t.Fatalf("expected updated word to be detected")
	}

	if err := TestSensitiveWordService.DeleteWord(t.Context(), word.ID); err != nil {
		t.Fatalf("DeleteWord failed: %v", err)
	}

	resp, err = TestSensitiveWordService.Check("这里有测试脏词B")
	if err != nil {
		t.Fatalf("Check after delete failed: %v", err)
	}
	if resp.Contain {
		t.Fatalf("expected deleted word to be removed from detector")
	}
}

func BenchmarkCheckSensitiveWords(b *testing.B) {
	for b.Loop() {
		TestSensitiveWordService.Check("你这个蠢猪真是个坏蛋")
	}
}
