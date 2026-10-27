package tests

import (
	"testing"
)

// Regresi: GORM melewati field ber-tag `default:` saat nilainya zero-value saat
// INSERT, sehingga DEFAULT database menimpa nilai yang diminta user.
//
// Dampaknya dulu: `remind_d_minus: 0` (mode hari-H) tersimpan 1, dan
// `is_active: false` (libur) tersimpan true. Fitur libur total maupun mode
// hari-H tidak bisa dipakai sama sekali lewat endpoint create.
//
// Test ini menjaga tag `default:` tidak dikembalikan diam-diam ke model.
func TestZeroValueFieldsAreNotOverriddenByDatabaseDefault(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	t.Run("remind_d_minus=0 (hari-H) harus tersimpan 0", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 1, "departure_time": "07:00", "remind_time": "05:00", "remind_d_minus": 0,
		})
		wantStatus(t, r, 200, "create hari-H")
		id := idOf(t, r, "create hari-H")

		if got := int8(r.data()["remind_d_minus"].(float64)); got != 0 {
			t.Errorf("response remind_d_minus = %d, want 0", got)
		}
		if got := dbInt(t, "destination_schedules", id, "remind_d_minus"); got != 0 {
			t.Errorf("kolom remind_d_minus = %d, want 0 (DEFAULT 1 menimpa)", got)
		}
	})

	t.Run("schedule is_active=false harus tersimpan false", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 2, "departure_time": "07:00", "remind_time": "20:00", "is_active": false,
		})
		wantStatus(t, r, 200, "create hari libur")
		id := idOf(t, r, "create hari libur")

		if r.data()["is_active"] != false {
			t.Error("response is_active = true, want false")
		}
		if got := dbInt(t, "destination_schedules", id, "is_active"); got != 0 {
			t.Errorf("kolom is_active = %d, want 0 (DEFAULT 1 menimpa)", got)
		}
	})

	t.Run("destination is_active=false harus tersimpan false", func(t *testing.T) {
		r := call(t, "POST", "/api/destination", token, map[string]interface{}{
			"name": "Liburan", "is_active": false,
		})
		wantStatus(t, r, 200, "create destinasi libur")
		id := idOf(t, r, "create destinasi libur")

		if got := dbInt(t, "destinations", id, "is_active"); got != 0 {
			t.Errorf("kolom destinations.is_active = %d, want 0 (DEFAULT 1 menimpa)", got)
		}
	})

	t.Run("checklist is_required=false harus tersimpan false", func(t *testing.T) {
		r := call(t, "POST", nested("/api/destination", desID, "/items"), token, map[string]interface{}{
			"name": "Barang Optional", "is_required": false,
		})
		wantStatus(t, r, 200, "create item optional")
		id := idOf(t, r, "create item optional")

		if got := dbInt(t, "checklist_items", id, "is_required"); got != 0 {
			t.Errorf("kolom is_required = %d, want 0", got)
		}
	})

	// Nilai non-zero tetap harus tersimpan benar (tidak ketiban default).
	t.Run("nilai non-default tetap tersimpan benar", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 3, "departure_time": "07:00", "remind_time": "20:00",
			"remind_d_minus": 1, "is_active": true,
		})
		wantStatus(t, r, 200, "create H-1")
		id := idOf(t, r, "create H-1")
		if got := dbInt(t, "destination_schedules", id, "remind_d_minus"); got != 1 {
			t.Errorf("remind_d_minus = %d, want 1", got)
		}
		if got := dbInt(t, "destination_schedules", id, "is_active"); got != 1 {
			t.Errorf("is_active = %d, want 1", got)
		}
	})
}

// Libur total (destination is_active=false) harus benar-benar menghentikan
// pembuatan reminder, kalau tidak fitur "libur" cuma jadi hiasan di response.
func TestInactiveDestinationProducesNoReminder(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	active := makeDestination(t, token, "Sekolah")
	libur := makeDestination(t, token, "Liburan")

	makeAllDaysSchedule(t, token, active)
	makeAllDaysSchedule(t, token, libur)

	// Matikan salah satu lewat PATCH (jalur update).
	r := call(t, "PATCH", "/api/destination/"+u64s(libur), token, map[string]interface{}{"is_active": false})
	wantStatus(t, r, 200, "matikan liburan")
	if r.data()["is_active"] != false {
		t.Fatal("destination libur tidak jadi is_active=false")
	}

	generateReminders(t)

	if got := countRows(t, "reminders", "destination_id = ?", active); got == 0 {
		t.Error("destinasi aktif tidak menghasilkan reminder")
	}
	if got := countRows(t, "reminders", "destination_id = ?", libur); got != 0 {
		t.Errorf("destinasi libur menghasilkan %d reminder, want 0", got)
	}
}

// Librahan per-hari (schedule is_active=false) juga harus dilewati worker.
func TestInactiveScheduleDayProducesNoReminder(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	total := countRows(t, "reminders", "destination_id = ?", desID)
	if total == 0 {
		t.Fatal("tidak ada reminder awal")
	}

	// Matikan semua jadwal.
	r := call(t, "PUT", base, token, map[string]interface{}{
		"schedules": inactiveAllDaysPayload(),
	})
	wantStatus(t, r, 200, "matikan semua hari")
	for _, row := range r.list() {
		m := row.(map[string]interface{})
		if m["is_active"] != false {
			t.Errorf("salah satu jadwal tidak jadi is_active=false")
		}
	}

	generateReminders(t)
	if got := countRows(t, "reminders", "destination_id = ? AND status = ?", desID, "pending"); got != 0 {
		t.Errorf("masih ada %d reminder pending padahal semua hari libur", got)
	}
}

func inactiveAllDaysPayload() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 7)
	for day := 0; day <= 6; day++ {
		out = append(out, map[string]interface{}{
			"day_of_week": day, "departure_time": "07:00",
			"remind_d_minus": 1, "remind_time": "20:00", "is_active": false,
		})
	}
	return out
}

func dbInt(t *testing.T, table string, id uint64, column string) int64 {
	t.Helper()
	var v int64
	if err := db().Table(table).Where("id = ?", id).Select(column).Scan(&v).Error; err != nil {
		t.Fatal(err)
	}
	return v
}
