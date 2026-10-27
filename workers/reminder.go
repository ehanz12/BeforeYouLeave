package workers

import (
	"context"
	"log"
	"time"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/services"
)

// StartReminderWorker = nyalakan dua loop di background:
//
//   - generator: buat reminder baru dari jadwal yang akan datang (jarang, default 60 menit)
//   - dispatcher: pending -> sent begitu remind_at-nya lewat (sering, default 5 menit)
//
// Dijalankan sekali dari main.go. Error tidak mematikan worker, hanya di-log,
// supaya satu siklus gagal tidak menghentikan siklus berikutnya.
func StartReminderWorker(ctx context.Context) {
	generateEvery := time.Duration(configs.AppConfig.ReminderGenerateIntervalMin) * time.Minute
	dispatchEvery := time.Duration(configs.AppConfig.ReminderDispatchIntervalMin) * time.Minute

	log.Printf("⏰ reminder worker aktif (generate=%dm, dispatch=%dm, horizon=%d hari)",
		configs.AppConfig.ReminderGenerateIntervalMin,
		configs.AppConfig.ReminderDispatchIntervalMin,
		configs.AppConfig.ReminderHorizonDays,
	)

	// Jalan sekali di startup supaya server yang baru nyala langsung punya reminder,
	// tidak perlu nunggu satu tick penuh.
	go runGenerate(ctx)
	go runDispatch(ctx)

	go loop(ctx, generateEvery, runGenerate)
	go loop(ctx, dispatchEvery, runDispatch)
}

func loop(ctx context.Context, every time.Duration, job func(context.Context)) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("⏹️ reminder worker dihentikan")
			return
		case <-ticker.C:
			job(ctx)
		}
	}
}

func runGenerate(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}
	count, err := services.GenerateReminders()
	if err != nil {
		log.Printf("⚠️ gagal generate reminder: %v", err)
		return
	}
	if count > 0 {
		log.Printf("✅ %d reminder baru dibuat", count)
	}
}

func runDispatch(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}
	count, err := services.DispatchDueReminders()
	if err != nil {
		log.Printf("⚠️ gagal dispatch reminder: %v", err)
		return
	}
	if count > 0 {
		log.Printf("🔔 %d reminder ditandai sent", count)
	}
}
