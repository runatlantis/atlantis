package ut

import (
	"testing"

	"github.com/go-playground/locales/en"
)

func BenchmarkBasicTranslation(b *testing.B) {

	en := en.New()
	ut := New(en, en)
	loc, found := ut.FindTranslator("en")
	if !found {
		b.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	translations := []struct {
		key      interface{}
		trans    string
		expected error
		override bool
	}{
		{
			key:      "welcome",
			trans:    "Welcome to the site",
			expected: nil,
		},
		{
			key:      "welcome-user",
			trans:    "Welcome to the site {0}",
			expected: nil,
		},
		{
			key:      "welcome-user2",
			trans:    "Welcome to the site {0}, your location is {1}",
			expected: nil,
		},
	}

	for _, tt := range translations {
		if err := loc.Add(tt.key, tt.trans, tt.override); err != nil {
			b.Fatalf("adding translation '%s' failed with key '%s'", tt.trans, tt.key)
		}
	}

	var err error

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = loc.T("welcome"); err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				if _, err = loc.T("welcome"); err != nil {
					b.Error(err)
				}
			}
		})
	})

	b.Run("With1Param", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = loc.T("welcome-user", "Joeybloggs"); err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("ParallelWith1Param", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				if _, err = loc.T("welcome-user", "Joeybloggs"); err != nil {
					b.Error(err)
				}
			}
		})
	})

	b.Run("With2Param", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = loc.T("welcome-user2", "Joeybloggs", "/dev/tty0"); err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("ParallelWith2Param", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				if _, err = loc.T("welcome-user2", "Joeybloggs", "/dev/tty0"); err != nil {
					b.Error(err)
				}
			}
		})
	})
}
