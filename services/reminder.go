package services

import (
	"errors"
	"time"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// appNow = waktu sekarang dalam timezone aplikasi.
// Harus sama dengan loc= di DSN database.go, kalau tidak H-1 vs hari-H bergeser.
func appNow() time.Time {
	return time.Now().In(configs.AppConfig.Location)
}

// startOfDay = buang jam-menit-detik, dipakai untuk kolom DATE remind_for_date.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

var validStatuses = map[string]bool{
	models.ReminderPending: true,
	models.ReminderSent:    true,
	models.ReminderRead:    true,
	models.ReminderSkipped: true,
}

// reminderMoment = hitung kapan reminder sebuah jadwal harus bunyi.
//
// remind_for_date = tanggal keberangkatan (misal Senin 2026-10-06).
// remind_at      = tanggal itu dikurangi 1 hari kalau remind_d_minus=1,
// lalu jamnya diganti dengan remind_time.
// Kalau hasilnya sudah lewat dari sekarang, reminder tidak perlu dibuat.
func reminderMoment(schedule models.DestinationSchedule, tripDate, now time.Time) (time.Time, error) {
	remindTime, err := time.Parse("15:04:05", schedule.RemindTime)
	if err != nil {
		return time.Time{}, err
	}

	remindDate := tripDate
	if schedule.RemindDMinus == models.RemindHMinus1 {
		remindDate = tripDate.AddDate(0, 0, -1)
	}

	moment := time.Date(
		remindDate.Year(), remindDate.Month(), remindDate.Day(),
		remindTime.Hour(), remindTime.Minute(), remindTime.Second(), 0,
		now.Location(),
	)
	if !moment.After(now) {
		return time.Time{}, nil
	}
	return moment, nil
}

// buildReminders = susun daftar reminder dari semua jadwal aktif.
//
// Destination is_active=FALSE (libur) atau schedule is_active=FALSE (libur per hari)
// dilewati, sesuai desain di migration destination_schedules.
func buildReminders(schedules []models.DestinationSchedule, now time.Time, horizonDays int) []models.Reminder {
	today := startOfDay(now)
	out := make([]models.Reminder, 0, len(schedules))

	for _, schedule := range schedules {
		if !schedule.IsActive || !schedule.Destination.IsActive {
			continue
		}

		for offset := 0; offset <= horizonDays; offset++ {
			tripDate := today.AddDate(0, 0, offset)
			if int(tripDate.Weekday()) != int(schedule.DayOfWeek) {
				continue
			}

			moment, err := reminderMoment(schedule, tripDate, now)
			if err != nil || moment.IsZero() {
				continue
			}

			out = append(out, models.Reminder{
				UserID:        schedule.Destination.UserID,
				DestinationID: schedule.DestinationID,
				ScheduleID:    schedule.ID,
				RemindForDate: tripDate,
				RemindAt:      moment,
				Status:        models.ReminderPending,
			})
		}
	}
	return out
}

// GenerateReminders = worker utama: buat reminder untuk jadwal yang akan datang.
//
// Aman dipanggil berulang: insert pakai OnConflict DoNothing, sehingga
// uq_reminders_schedule_for_date (schedule_id, remind_for_date) yang mencegah dobel.
// Return jumlah reminder yang benar-benar dibuat.
func GenerateReminders() (int, error) {
	now := appNow()
	horizon := configs.AppConfig.ReminderHorizonDays

	var schedules []models.DestinationSchedule
	if err := databases.DB.
		Preload("Destination").
		Joins("JOIN destinations ON destinations.id = destination_schedules.destination_id").
		Where("destination_schedules.is_active = ?", true).
		Where("destinations.is_active = ?", true).
		Find(&schedules).Error; err != nil {
		return 0, err
	}

	candidates := buildReminders(schedules, now, horizon)
	if len(candidates) == 0 {
		return 0, nil
	}

	if err := databases.DB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "schedule_id"}, {Name: "remind_for_date"}},
			DoNothing: true,
		}).
		Create(&candidates).Error; err != nil {
		return 0, err
	}

	items, err := backfillItemChecks(startOfDay(now), startOfDay(now).AddDate(0, 0, horizon))
	if err != nil {
		return 0, err
	}
	return items, nil
}

// backfillItemChecks = copy checklist_items destination ke reminder_item_checks.
//
// Dijalankan untuk reminder dalam rentang horizon yang belum punya satupun item check.
// Selain untuk reminder baru, ini juga membuat reminder lama ikut terisi kalau user
// menambah barang baru ke checklist-nya setelah reminder dibuat.
func backfillItemChecks(from, to time.Time) (int, error) {
	var reminders []models.Reminder
	if err := databases.DB.
		Preload("ItemChecks").
		Where("remind_for_date >= ? AND remind_for_date <= ?", from, to).
		Where("status IN ?", []string{models.ReminderPending, models.ReminderSent}).
		Find(&reminders).Error; err != nil {
		return 0, err
	}

	created := 0
	for i := range reminders {
		if len(reminders[i].ItemChecks) > 0 {
			continue
		}

		var items []models.ChecklistItem
		if err := databases.DB.
			Where("destination_id = ?", reminders[i].DestinationID).
			Find(&items).Error; err != nil {
			return created, err
		}
		// Destinasi tanpa barang: tidak ada yang dicopy, dan tidak diulang tiap run.
		if len(items) == 0 {
			continue
		}

		checks := make([]models.ReminderItemCheck, 0, len(items))
		for _, item := range items {
			checks = append(checks, models.ReminderItemCheck{
				ReminderID:      reminders[i].ID,
				ChecklistItemID: item.ID,
				IsChecked:       false,
			})
		}

		if err := databases.DB.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "reminder_id"}, {Name: "checklist_item_id"}},
				DoNothing: true,
			}).
			Create(&checks).Error; err != nil {
			return created, err
		}
		created += len(checks)
	}
	return created, nil
}

// DispatchDueReminders = pending -> sent untuk reminder yang remind_at-nya sudah lewat.
//
// Status sent berarti "sudah bunyi tapi belum dibaca", jadi FE bisa bunyi ulang /
// tampil sticky sampai user menandainya baca (lihat Reminder.StillRinging).
func DispatchDueReminders() (int64, error) {
	now := appNow()
	result := databases.DB.Model(&models.Reminder{}).
		Where("status = ?", models.ReminderPending).
		Where("remind_at <= ?", now).
		Updates(map[string]interface{}{
			"status":  models.ReminderSent,
			"sent_at": now,
		})
	return result.RowsAffected, result.Error
}

// getOwnedReminder = reminder harus milik user yang sedang login.
func getOwnedReminder(reminderID, userID uint64) (models.Reminder, error) {
	var r models.Reminder
	if err := databases.DB.
		Where("id = ? AND user_id = ?", reminderID, userID).
		First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Reminder{}, errors.New("reminder tidak ditemukan !")
		}
		return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
	}
	return r, nil
}

func ListReminders(req request.ListReminderRequest, userID uint64, page, limit int) ([]models.Reminder, int64, error) {
	query := databases.DB.Model(&models.Reminder{}).Where("user_id = ?", userID)

	switch {
	case req.UnreadOnly:
		query = query.Where("status IN ?", []string{models.ReminderPending, models.ReminderSent})
	case req.Status != "":
		if !validStatuses[req.Status] {
			return nil, 0, errors.New("status tidak valid (pending/sent/read/skipped) !")
		}
		query = query.Where("status = ?", req.Status)
	}

	if req.DestinationID > 0 {
		query = query.Where("destination_id = ?", req.DestinationID)
	}
	if req.StartDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, errors.New("format start_date tidak valid (YYYY-MM-DD) !")
		}
		query = query.Where("remind_for_date >= ?", startOfDay(start.In(configs.AppConfig.Location)))
	}
	if req.EndDate != "" {
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, errors.New("format end_date tidak valid (YYYY-MM-DD) !")
		}
		query = query.Where("remind_for_date <= ?", startOfDay(end.In(configs.AppConfig.Location)))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}

	var reminders []models.Reminder
	offset := (page - 1) * limit
	if err := query.
		Preload("Destination").
		Order("remind_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&reminders).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}
	return reminders, total, nil
}

func DetailReminder(reminderID, userID uint64) (models.Reminder, error) {
	if _, err := getOwnedReminder(reminderID, userID); err != nil {
		return models.Reminder{}, err
	}

	var r models.Reminder
	if err := databases.DB.
		Preload("Destination").
		Preload("ItemChecks", func(db *gorm.DB) *gorm.DB {
			return db.Order("reminder_item_checks.checklist_item_id ASC")
		}).
		Preload("ItemChecks.ChecklistItem").
		Where("id = ? AND user_id = ?", reminderID, userID).
		First(&r).Error; err != nil {
		return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
	}
	return r, nil
}

// MarkReminderRead = cara mematikan reminder adalah menandainya sudah dibaca.
func MarkReminderRead(reminderID, userID uint64) (models.Reminder, error) {
	r, err := getOwnedReminder(reminderID, userID)
	if err != nil {
		return models.Reminder{}, err
	}

	// Sudah read = idempotent, jangan geser read_at ke waktu sekarang.
	if r.Status != models.ReminderRead {
		r.MarkAsRead()
		if err := databases.DB.Model(&r).Updates(map[string]interface{}{
			"status":  r.Status,
			"read_at": r.ReadAt,
		}).Error; err != nil {
			return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
		}
	}
	return DetailReminder(reminderID, userID)
}

// SkipReminder = batalkan reminder ini saja (misal libur mendadak),
// tanpa menyentuh jadwalnya. Cara ini yang dipakai user, bukan liburkan jadwal.
func SkipReminder(reminderID, userID uint64) (models.Reminder, error) {
	r, err := getOwnedReminder(reminderID, userID)
	if err != nil {
		return models.Reminder{}, err
	}

	if err := databases.DB.Model(&r).Update("status", models.ReminderSkipped).Error; err != nil {
		return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
	}
	return DetailReminder(reminderID, userID)
}

// UpdateItemCheck = centang / uncentang satu barang pada sebuah reminder.
//
// Baris item check dibuat kalau belum ada (barang ditambahkan setelah reminder
// dibuat), jadi endpoint ini tidak bergantung penuh pada worker backfill.
func UpdateItemCheck(req request.UpdateItemCheckRequest, reminderID, checklistItemID, userID uint64) (models.Reminder, error) {
	r, err := getOwnedReminder(reminderID, userID)
	if err != nil {
		return models.Reminder{}, err
	}

	// Barang harus milik destination milik user itu juga, supaya tidak bisa
	// menempelkan checklist item destination lain ke reminder ini.
	var item models.ChecklistItem
	if err := databases.DB.
		Where("id = ? AND destination_id = ?", checklistItemID, r.DestinationID).
		First(&item).Error; err != nil {
		return models.Reminder{}, errors.New("item tidak ditemukan !")
	}

	seed := models.ReminderItemCheck{
		ReminderID:      reminderID,
		ChecklistItemID: checklistItemID,
	}
	if err := databases.DB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "reminder_id"}, {Name: "checklist_item_id"}},
			DoNothing: true,
		}).
		Create(&seed).Error; err != nil {
		return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
	}

	checkedAt := interface{}(nil)
	if req.IsChecked {
		checkedAt = appNow()
	}

	if err := databases.DB.Model(&models.ReminderItemCheck{}).
		Where("reminder_id = ? AND checklist_item_id = ?", reminderID, checklistItemID).
		Updates(map[string]interface{}{
			"is_checked": req.IsChecked,
			"checked_at": checkedAt,
		}).Error; err != nil {
		return models.Reminder{}, errors.New("terjadi kesalahan sistem !")
	}

	return DetailReminder(reminderID, userID)
}
