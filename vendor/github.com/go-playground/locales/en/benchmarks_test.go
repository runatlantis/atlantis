package en

import (
	"testing"
	"time"

	"github.com/go-playground/locales/currency"
)

func BenchmarkFmtNumber(b *testing.B) {

	trans := New()
	f64 := float64(1234567.43)
	precision := uint64(2)

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtNumber(f64, precision)
		}
	})

	b.Run("Parallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtNumber(f64, precision)
			}
		})
	})
}

func BenchmarkFmtPercent(b *testing.B) {

	trans := New()
	f64 := float64(97.05)
	precision := uint64(2)

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtPercent(f64, precision)
		}
	})

	b.Run("Parallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtPercent(f64, precision)
			}
		})
	})
}

func BenchmarkFmtCurrency(b *testing.B) {

	trans := New()
	f64 := float64(1234567.43)
	precision := uint64(2)

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtCurrency(f64, precision, currency.USD)
		}
	})

	b.Run("Parallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtCurrency(f64, precision, currency.USD)
			}
		})
	})
}

func BenchmarkFmtAccounting(b *testing.B) {

	trans := New()
	f64 := float64(1234567.43)
	precision := uint64(2)

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtAccounting(f64, precision, currency.USD)
		}
	})

	b.Run("Parallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtAccounting(f64, precision, currency.USD)
			}
		})
	})
}

func BenchmarkFmtDate(b *testing.B) {

	trans := New()
	t := time.Now()

	b.ResetTimer()

	b.Run("FmtDateShort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtDateShort(t)
		}
	})

	b.Run("FmtDateShortParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtDateShort(t)
			}
		})
	})

	b.Run("FmtDateMedium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtDateMedium(t)
		}
	})

	b.Run("FmtDateMediumParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtDateMedium(t)
			}
		})
	})

	b.Run("FmtDateLong", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtDateLong(t)
		}
	})

	b.Run("FmtDateLongParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtDateLong(t)
			}
		})
	})

	b.Run("FmtDateFull", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtDateFull(t)
		}
	})

	b.Run("FmtDateFullParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtDateFull(t)
			}
		})
	})
}

func BenchmarkFmtTime(b *testing.B) {

	trans := New()
	t := time.Now()

	b.ResetTimer()

	b.Run("FmtTimeShort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtTimeShort(t)
		}
	})

	b.Run("FmtTimeShortParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtTimeShort(t)
			}
		})
	})

	b.Run("FmtTimeMedium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtTimeMedium(t)
		}
	})

	b.Run("FmtTimeMediumParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtTimeMedium(t)
			}
		})
	})

	b.Run("FmtTimeLong", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtTimeLong(t)
		}
	})

	b.Run("FmtTimeLongParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtTimeLong(t)
			}
		})
	})

	b.Run("FmtTimeFull", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			trans.FmtTimeFull(t)
		}
	})

	b.Run("FmtTimeFullParallel", func(b *testing.B) {

		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				trans.FmtTimeFull(t)
			}
		})
	})
}
