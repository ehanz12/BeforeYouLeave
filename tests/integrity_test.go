package tests

import "testing"

// Hapus destination harus menyapu semua data turunannya lewat ON DELETE CASCADE.
// Kalau tidak, akan ada baris yatim yang menghasilkan reminder untuk kondisi
// yang sudah dihapus user.
func TestDeleteDestinationCascadesEverything(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeItem(t, token, desID, "Buku")
	makeItem(t, token, desID, "Pena")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	reminders := countRows(t, "reminders", "destination_id = ?", desID)
	checks := countRows(t, "reminder_item_checks", "")
	if reminders == 0 || checks == 0 {
		t.Fatal("data turunannya belum ada, test tidak bermakna")
	}

	r := call(t, "DELETE", "/api/destination/"+u64s(desID), token, nil)
	wantStatus(t, r, 200, "delete destination")

	checksAfter := []struct {
		table string
		where string
	}{
		{"destinations", "id = ?"},
		{"checklist_items", "destination_id = ?"},
		{"destination_schedules", "destination_id = ?"},
		{"reminders", "destination_id = ?"},
		{"reminder_item_checks", ""},
	}
	for _, c := range checksAfter {
		if c.where == "" {
			if got := countRows(t, c.table, ""); got != 0 {
				t.Errorf("%s = %d baris, want 0 (yatim)", c.table, got)
			}
			continue
		}
		if got := countRows(t, c.table, c.where, desID); got != 0 {
			t.Errorf("%s = %d baris, want 0 (tidak tercascade)", c.table, got)
		}
	}
	// User-nya sendiri harus tetap ada.
	if countRows(t, "users", "") != 1 {
		t.Error("user ikut terhapus")
	}
}

// Destinasi lain milik user yang sama tidak boleh ikut terhapus.
func TestDeleteDestinationKeepsSiblings(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	sekolah := makeDestination(t, token, "Sekolah")
	kantor := makeDestination(t, token, "Kantor")
	makeItem(t, token, sekolah, "Buku")
	makeItem(t, token, kantor, "Laptop")

	call(t, "DELETE", "/api/destination/"+u64s(sekolah), token, nil)

	if countRows(t, "destinations", "id = ?", kantor) != 1 {
		t.Error("destinasi lain ikut terhapus")
	}
	if countRows(t, "checklist_items", "destination_id = ?", kantor) != 1 {
		t.Error("checklist item kantor ikut terhapus")
	}
	if countRows(t, "checklist_items", "destination_id = ?", sekolah) != 0 {
		t.Error("checklist items sekolah tidak terhapus")
	}
}

// Hapus satu hari jadwal harus menghapus reminder yang berasal dari hari itu,
// tapi reminder dari hari lain tetap utuh.
func TestDeleteScheduleDayRemovesOnlyItsReminders(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	total := countRows(t, "reminders", "destination_id = ?", desID)
	if total == 0 {
		t.Fatal("tidak ada reminder")
	}

	var schID uint64
	db().Table("destination_schedules").
		Where("destination_id = ? AND day_of_week = ?", desID, 1).Select("id").Scan(&schID)
	if schID == 0 {
		t.Fatal("jadwal Senin tidak ditemukan")
	}
	before := countRows(t, "reminders", "schedule_id = ?", schID)
	if before == 0 {
		t.Fatal("tidak ada reminder milik jadwal Senin")
	}

	r := call(t, "DELETE", base+"/"+u64s(schID), token, nil)
	wantStatus(t, r, 200, "delete jadwal Senin")

	if got := countRows(t, "reminders", "schedule_id = ?", schID); got != 0 {
		t.Errorf("reminder milik jadwal yang dihapus = %d, want 0", got)
	}
	if got := countRows(t, "reminders", "destination_id = ?", desID); got != total-before {
		t.Errorf("total reminder = %d, want %d (hanya reminder Senin yang hilang)", got, total-before)
	}
	if got := countRows(t, "destination_schedules", "destination_id = ?", desID); got != 6 {
		t.Errorf("sisa jadwal = %d, want 6", got)
	}
}

// Integrity check menyeluruh: tidak boleh ada baris yang menunjuk ke induk yang hilang.
func TestNoOrphanRows(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	des := makeDestination(t, token, "Sekolah")
	makeItem(t, token, des, "Buku")
	makeAllDaysSchedule(t, token, des)
	generateReminders(t)

	orphans := []struct {
		name  string
		query string
	}{
		{"checklist_items -> destination",
			"SELECT COUNT(*) FROM checklist_items ci LEFT JOIN destinations d ON d.id = ci.destination_id WHERE d.id IS NULL"},
		{"destination_schedules -> destination",
			"SELECT COUNT(*) FROM destination_schedules s LEFT JOIN destinations d ON d.id = s.destination_id WHERE d.id IS NULL"},
		{"reminders -> destination",
			"SELECT COUNT(*) FROM reminders r LEFT JOIN destinations d ON d.id = r.destination_id WHERE d.id IS NULL"},
		{"reminders -> schedule",
			"SELECT COUNT(*) FROM reminders r LEFT JOIN destination_schedules s ON s.id = r.schedule_id WHERE s.id IS NULL"},
		{"reminders -> user",
			"SELECT COUNT(*) FROM reminders r LEFT JOIN users u ON u.id = r.user_id WHERE u.id IS NULL"},
		{"reminder_item_checks -> reminder",
			"SELECT COUNT(*) FROM reminder_item_checks ric LEFT JOIN reminders r ON r.id = ric.reminder_id WHERE r.id IS NULL"},
		{"reminder_item_checks -> checklist_item",
			"SELECT COUNT(*) FROM reminder_item_checks ric LEFT JOIN checklist_items ci ON ci.id = ric.checklist_item_id WHERE ci.id IS NULL"},
	}
	for _, o := range orphans {
		var n int64
		db().Raw(o.query).Scan(&n)
		if n != 0 {
			t.Errorf("%s: %d baris yatim", o.name, n)
		}
	}
}

// User dihapus harus menyapu semua destination dan turunannya.
func TestDeleteUserCascadesAll(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	des := makeDestination(t, token, "Sekolah")
	makeItem(t, token, des, "Buku")
	makeAllDaysSchedule(t, token, des)
	generateReminders(t)

	if err := db().Exec("DELETE FROM users").Error; err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"destinations", "checklist_items", "destination_schedules", "reminders", "reminder_item_checks"} {
		if got := countRows(t, tbl, ""); got != 0 {
			t.Errorf("%s = %d, want 0 setelah user dihapus", tbl, got)
		}
	}
}
