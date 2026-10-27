package tests

import (
	"testing"
	"time"
)

// Horizon di env test = 14. Jadwal aktif setiap hari (0..6) dengan H-1 23:00:
// offset 0 (hari ini) selalu ter-skip karena remind_at-nya sudah lewat,
// jadi hasil yang diharapkan = 14.
func TestGenerateRemindersCount(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)

	generateReminders(t)

	if got := countRows(t, "reminders", "destination_id = ?", desID); got != 14 {
		t.Errorf("reminder = %d, want 14 (14 hari horizon, hari ini ter-skip)", got)
	}
	if got := countRows(t, "reminders", "status = ?", "pending"); got != 14 {
		t.Errorf("semua harus pending, dapat %d pending", got)
	}
}

// Uq_reminders_schedule_for_date harus mencegah dobel walau worker jalan
// berulang (crontab, restart, dsb).
func TestGenerateRemindersIsIdempotent(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)

	generateReminders(t)
	first := countRows(t, "reminders", "")

	for i := 0; i < 5; i++ {
		generateReminders(t)
	}

	if got := countRows(t, "reminders", ""); got != first {
		t.Errorf("setelah 6x generate: %d reminder, want %d (ada duplikat)", got, first)
	}

	// Tidak boleh ada pasangan (schedule_id, remind_for_date) yang sama.
	var dupes int64
	db().Raw(`
		SELECT COUNT(*) FROM (
			SELECT schedule_id, remind_for_date
			FROM reminders
			GROUP BY schedule_id, remind_for_date
			HAVING COUNT(*) > 1
		) d`).Scan(&dupes)
	if dupes != 0 {
		t.Errorf("ada %d pasangan jadwal+tanggal yang duplikat", dupes)
	}
}

// Backfill: checklist_items destination harus dicopy ke reminder_item_checks.
func TestGenerateRemindersCopiesChecklistItems(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")

	for _, name := range []string{"Buku", "Pena", "Botol"} {
		makeItem(t, token, desID, name)
	}
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	reminders := countRows(t, "reminders", "destination_id = ?", desID)
	if reminders == 0 {
		t.Fatal("tidak ada reminder")
	}
	// 14 reminder x 3 item = 42 baris.
	if got := countRows(t, "reminder_item_checks", ""); got != reminders*3 {
		t.Errorf("reminder_item_checks = %d, want %d (%d reminder x 3 item)", got, reminders*3, reminders)
	}
	if got := countRows(t, "reminder_item_checks", "is_checked = 1"); got != 0 {
		t.Errorf("%d item tercentang sejak awal, harusnya 0", got)
	}
	if got := countRows(t, "reminder_item_checks", "checked_at IS NOT NULL"); got != 0 {
		t.Errorf("%d item punya checked_at, harusnya 0", got)
	}

	// Item dari destination lain tidak boleh ikut ke reminder ini.
	other := makeDestination(t, token, "Kampus")
	makeItem(t, token, other, "Laptop")
	generateReminders(t)
	if got := countRows(t, "reminder_item_checks", ""); got != reminders*3 {
		t.Errorf("setelah tambah item di destination lain = %d, want %d", got, reminders*3)
	}
}

// Destinasi tanpa checklist: tidak boleh error, dan tidak boleh bikin reminder
// bolong yang nanti di-backfill ulang terus-menerus.
func TestGenerateRemindersWithoutChecklistItems(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)

	generateReminders(t)
	if countRows(t, "reminders", "") == 0 {
		t.Fatal("tidak ada reminder")
	}
	if got := countRows(t, "reminder_item_checks", ""); got != 0 {
		t.Errorf("reminder_item_checks = %d tanpa checklist, want 0", got)
	}
	// Jalankan lagi, harus tetap stabil (tidak Creating apa-apa).
	generateReminders(t)
	if got := countRows(t, "reminders", ""); got != 14 {
		t.Errorf("reminder = %d, want 14 (stabil)", got)
	}
}

// Dispatch: pending -> sent hanya untuk yang remind_at sudah lewat.
func TestDispatchDueReminders(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	// Semua masih di masa depan -> tidak ada yang boleh di-dispatch.
	if n := dispatchReminders(t); n != 0 {
		t.Errorf("dispatch = %d, want 0 (semua masih di masa depan)", n)
	}
	if got := countRows(t, "reminders", "status = ?", "pending"); got != 14 {
		t.Errorf("pending = %d, want 14", got)
	}

	// Mundurkan 1 reminder ke masa lalu.
	var id uint64
	db().Table("reminders").Where("status = ?", "pending").Order("remind_at ASC").Limit(1).Select("id").Scan(&id)
	if err := db().Exec("UPDATE reminders SET remind_at = ? WHERE id = ?",
		time.Now().Add(-2*time.Hour), id).Error; err != nil {
		t.Fatal(err)
	}

	if n := dispatchReminders(t); n != 1 {
		t.Errorf("dispatch = %d, want 1", n)
	}
	if got := countRows(t, "reminders", "id = ? AND status = ?", id, "sent"); got != 1 {
		t.Error("reminder yang jatuh tempo tidak jadi sent")
	}
	if got := countRows(t, "reminders", "id = ? AND sent_at IS NOT NULL", id); got != 1 {
		t.Error("sent_at tidak diisi")
	}
	if got := countRows(t, "reminders", "status = ?", "pending"); got != 13 {
		t.Errorf("pending = %d, want 13", got)
	}

	// Idempoten: dispatch kedua tidak mengubah apa pun.
	if n := dispatchReminders(t); n != 0 {
		t.Errorf("dispatch kedua = %d, want 0", n)
	}
}

// Dispatch tidak boleh menyentuh reminder yang sudah read atau skipped.
func TestDispatchDoesNotResurrectReadOrSkipped(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	past := time.Now().Add(-time.Hour)
	if err := db().Exec("UPDATE reminders SET remind_at = ? WHERE status = ?", past, "pending").Error; err != nil {
		t.Fatal(err)
	}
	if err := db().Exec("UPDATE reminders SET status = 'read' WHERE id = (SELECT x FROM (SELECT MIN(id) AS x FROM reminders) m)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db().Exec("UPDATE reminders SET status = 'skipped' WHERE id = (SELECT x FROM (SELECT MIN(id)+1 AS x FROM reminders) m)").Error; err != nil {
		t.Fatal(err)
	}

	dispatchReminders(t)

	if got := countRows(t, "reminders", "status = ?", "read"); got != 1 {
		t.Errorf("read = %d, want 1 (tidak boleh jadi sent lagi)", got)
	}
	if got := countRows(t, "reminders", "status = ?", "skipped"); got != 1 {
		t.Errorf("skipped = %d, want 1 (tidak boleh jadi sent lagi)", got)
	}
	if got := countRows(t, "reminders", "status = ?", "sent"); got != 12 {
		t.Errorf("sent = %d, want 12 (14 - 1 read - 1 skipped)", got)
	}
}

// Tanggal keberangkatan harus sesuai hari dalam seminggu dari jadwal, dan
// remind_at harus dihitung relatif terhadap remind_d_minus.
func TestGeneratedReminderDatesMatchScheduleDays(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")

	// Hanya hari Wednesday (3), H-1 23:00.
	r := call(t, "PUT", nested("/api/destination", desID, "/schedules"), token, map[string]interface{}{
		"schedules": []map[string]interface{}{
			{"day_of_week": 3, "departure_time": "07:00", "remind_time": "23:00", "remind_d_minus": 1},
		},
	})
	wantStatus(t, r, 200, "jadwal rabu")
	generateReminders(t)

	type row struct {
		RemindForDate time.Time
		RemindAt      time.Time
	}
	var rows []row
	if err := db().Table("reminders").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("tidak ada reminder untuk jadwal rabu")
	}
	for _, rw := range rows {
		if rw.RemindForDate.Weekday() != time.Wednesday {
			t.Errorf("remind_for_date %s hari=%v, want Rabu",
				rw.RemindForDate.Format("2006-01-02"), rw.RemindForDate.Weekday())
		}
		// H-1: bunyinya sehari sebelum, jam 23:00.
		wantAt := time.Date(
			rw.RemindForDate.Year(), rw.RemindForDate.Month(), rw.RemindForDate.Day()-1, 23, 0, 0, 0, rw.RemindAt.Location())
		if !rw.RemindAt.Equal(wantAt) {
			t.Errorf("remind_at = %s, want %s", rw.RemindAt.Format("2006-01-02 15:04"), wantAt.Format("2006-01-02 15:04"))
		}
		if !rw.RemindAt.After(time.Now()) {
			t.Errorf("reminder %s di masa lalu, tidak seharusnya dibuat", rw.RemindAt)
		}
	}
}

// Jadwal H-0: bunyinya di hari yang sama, sebelum berangkat.
func TestGeneratedReminderSameDay(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")

	r := call(t, "PUT", nested("/api/destination", desID, "/schedules"), token, map[string]interface{}{
		"schedules": []map[string]interface{}{
			{"day_of_week": 5, "departure_time": "23:59", "remind_time": "00:01", "remind_d_minus": 0},
		},
	})
	wantStatus(t, r, 200, "jadwal jumat H-0")
	generateReminders(t)

	type row struct {
		RemindForDate time.Time
		RemindAt      time.Time
	}
	var rows []row
	db().Table("reminders").Find(&rows)
	if len(rows) == 0 {
		t.Fatal("tidak ada reminder H-0")
	}
	for _, rw := range rows {
		if rw.RemindForDate.Weekday() != time.Friday {
			t.Errorf("remind_for_date %v bukan Jumat", rw.RemindForDate.Weekday())
		}
		sameDay := rw.RemindAt.Year() == rw.RemindForDate.Year() &&
			rw.RemindAt.YearDay() == rw.RemindForDate.YearDay()
		if !sameDay {
			t.Errorf("H-0 tapi bunyi di tanggal lain: %s", rw.RemindAt.Format("2006-01-02"))
		}
		if rw.RemindAt.Hour() != 0 || rw.RemindAt.Minute() != 1 {
			t.Errorf("remind_at = %s, want 00:01", rw.RemindAt.Format("15:04"))
		}
	}
}
