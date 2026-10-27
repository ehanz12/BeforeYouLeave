package tests

import (
	"strings"
	"testing"
)

func TestScheduleCreateAndValidation(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	t.Run("create valid", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1,
		})
		wantStatus(t, r, 200, "create schedule")
		// jam harus ternormalisasi jadi HH:MM:SS
		if got := r.data()["departure_time"]; got != "07:00:00" {
			t.Errorf("departure_time = %v, want 07:00:00", got)
		}
		if got := r.data()["remind_time"]; got != "20:00:00" {
			t.Errorf("remind_time = %v, want 20:00:00", got)
		}
		if got := r.data()["is_active"]; got != true {
			t.Error("is_active default = false, want true")
		}
		if got := uint64(r.data()["destination_id"].(float64)); got != desID {
			t.Errorf("destination_id = %d, want %d", got, desID)
		}
	})

	t.Run("hari sama dobel ditolak", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 1, "departure_time": "08:00", "remind_time": "21:00",
		})
		if r.Status == 200 {
			t.Error("hari dobel = 200, seharusnya ditolak (ada UNIQUE destination_id+day_of_week)")
		}
		if countRows(t, "destination_schedules", "destination_id = ? AND day_of_week = ?", desID, 1) != 1 {
			t.Error("harusnya tetap 1 baris untuk hari itu")
		}
	})

	t.Run("day_of_week di luar 0..6 ditolak", func(t *testing.T) {
		for _, day := range []int{-1, 7, 100} {
			r := call(t, "POST", base, token, map[string]interface{}{
				"day_of_week": day, "departure_time": "07:00", "remind_time": "20:00",
			})
			if r.Status == 200 {
				t.Errorf("day_of_week=%d = 200, seharusnya ditolak", day)
			}
		}
	})

	t.Run("jam ngawur ditolak", func(t *testing.T) {
		for _, tm := range []string{"", "24:00", "07:60", "abc", "07"} {
			r := call(t, "POST", base, token, map[string]interface{}{
				"day_of_week": 2, "departure_time": "07:00", "remind_time": tm,
			})
			if r.Status == 200 {
				t.Errorf("remind_time=%q = 200, seharusnya ditolak", tm)
			}
		}
	})

	t.Run("remind_d_minus ngawur ditolak", func(t *testing.T) {
		for _, d := range []int{-1, 2, 5} {
			r := call(t, "POST", base, token, map[string]interface{}{
				"day_of_week": 2, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": d,
			})
			if r.Status == 200 {
				t.Errorf("remind_d_minus=%d = 200, seharusnya ditolak", d)
			}
		}
	})

	t.Run("hari-H: remind harus sebelum berangkat", func(t *testing.T) {
		// remind 20:00 sementara berangkat 07:00 -> tidak masuk akal.
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 2, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 0,
		})
		if r.Status == 200 {
			t.Error("reminder hari-H setelah berangkat = 200, seharusnya ditolak")
		}
		// remind 05:00 sebelum berangkat 07:00 -> valid.
		r = call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 2, "departure_time": "07:00", "remind_time": "05:00", "remind_d_minus": 0,
		})
		wantStatus(t, r, 200, "hari-H remind sebelum berangkat")
	})

	t.Run("H-1 tidak punya aturan urutan jam", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 3, "departure_time": "07:00", "remind_time": "22:00", "remind_d_minus": 1,
		})
		wantStatus(t, r, 200, "H-1 dengan remind 22:00 harusnya valid")
	})
}

func TestScheduleUpdate(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	r := call(t, "POST", base, token, map[string]interface{}{
		"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1,
	})
	id := idOf(t, r, "create")

	t.Run("update is_active saja", func(t *testing.T) {
		r := call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{"is_active": false})
		wantStatus(t, r, 200, "toggle is_active")
		if r.data()["is_active"] != false {
			t.Error("is_active tidak jadi false")
		}
		if r.data()["departure_time"] != "07:00:00" {
			t.Error("field lain ikut berubah")
		}
	})

	t.Run("update hari ke hari yang sudah dipakai ditolak", func(t *testing.T) {
		call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 2, "departure_time": "07:00", "remind_time": "20:00",
		})
		r := call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{"day_of_week": 2})
		if r.Status == 200 {
			t.Error("pindah ke hari yang terpakai = 200, seharusnya ditolak")
		}
	})

	// Bug hunt: update hanya remind_time harus tetap divalidasi terhadap departure_time.
	t.Run("update remind_time jadi setelah berangkat ditolak", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": 4, "departure_time": "07:00", "remind_time": "05:00", "remind_d_minus": 0,
		})
		sameDay := idOf(t, r, "create same-day")

		// Kalau hanya remind_time yang dikirim, validasi harus tetap berjalan
		// memakai departure_time yang sudah tersimpan (07:00).
		r = call(t, "PATCH", base+"/"+u64s(sameDay), token, map[string]interface{}{"remind_time": "20:00"})
		if r.Status == 200 {
			t.Error("remind_time 20:00 di jadwal hari-H (berangkat 07:00) = 200, seharusnya ditolak")
		}
	})

	t.Run("patch kosong ditolak", func(t *testing.T) {
		r := call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{})
		if r.Status == 200 {
			t.Error("patch kosong = 200, seharusnya ditolak")
		}
	})

	t.Run("hapus jadwal", func(t *testing.T) {
		r := call(t, "DELETE", base+"/"+u64s(id), token, nil)
		wantStatus(t, r, 200, "delete schedule")
		if countRows(t, "destination_schedules", "id = ?", id) != 0 {
			t.Error("jadwal masih ada setelah delete")
		}
	})
}

func TestScheduleBulkReplace(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	t.Run("ganti dari kosong ke 5 hari", func(t *testing.T) {
		r := call(t, "PUT", base, token, map[string]interface{}{
			"schedules": []map[string]interface{}{
				{"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1},
				{"day_of_week": 2, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1},
				{"day_of_week": 3, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1},
				{"day_of_week": 4, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1},
				{"day_of_week": 5, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1},
			},
		})
		wantStatus(t, r, 200, "bulk 5 hari")
		if n := len(r.list()); n != 5 {
			t.Errorf("tersimpan = %d, want 5", n)
		}
		if countRows(t, "destination_schedules", "destination_id = ?", desID) != 5 {
			t.Error("jumlah baris di DB tidak 5")
		}
	})

	t.Run("replaceall menghapus hari lama", func(t *testing.T) {
		// Sekarang hanya Jumat-Sabtu.
		r := call(t, "PUT", base, token, map[string]interface{}{
			"schedules": []map[string]interface{}{
				{"day_of_week": 5, "departure_time": "06:30", "remind_time": "19:00"},
				{"day_of_week": 6, "departure_time": "08:00", "remind_time": "21:00"},
			},
		})
		wantStatus(t, r, 200, "replace 2 hari")
		if countRows(t, "destination_schedules", "destination_id = ?", desID) != 2 {
			t.Error("replaceall tidak menimpa, masih ada baris lama")
		}
		// Response harus urut hari (Minggu..Sabtu) -> Jumat lalu Sabtu
		list := r.list()
		if len(list) != 2 {
			t.Fatalf("response berisi %d jadwal, want 2", len(list))
		}
		first, _ := list[0].(map[string]interface{})
		if got := first["day_of_week"]; got != float64(5) {
			t.Errorf("urutan response salah, awal = %v, want 5", got)
		}
	})

	t.Run("payload kosong ditolak", func(t *testing.T) {
		r := call(t, "PUT", base, token, map[string]interface{}{"schedules": []map[string]interface{}{}})
		if r.Status == 200 {
			t.Error("schedules kosong = 200, seharusnya ditolak")
		}
	})

	// Yang penting: transaksi harus all-or-nothing.
	t.Run("salah satu baris invalid = tidak ada yang tersimpan", func(t *testing.T) {
		before := countRows(t, "destination_schedules", "destination_id = ?", desID)

		r := call(t, "PUT", base, token, map[string]interface{}{
			"schedules": []map[string]interface{}{
				{"day_of_week": 0, "departure_time": "07:00", "remind_time": "20:00"},
				{"day_of_week": 9, "departure_time": "07:00", "remind_time": "20:00"},
			},
		})
		if r.Status == 200 {
			t.Error("payload dengan day_of_week=9 = 200, seharusnya ditolak")
		}
		after := countRows(t, "destination_schedules", "destination_id = ?", desID)
		if after != before {
			t.Errorf("baris berubah %d -> %d, transaksi seharusnya rollback penuh", before, after)
		}
		if r.errMsg() == "" || !strings.Contains(r.errMsg(), "schedules[1]") {
			t.Errorf("pesan error tidak menunjuk indeks baris: %q", r.errMsg())
		}
	})

	t.Run("hari dobel di dalam payload ditolak", func(t *testing.T) {
		before := countRows(t, "destination_schedules", "destination_id = ?", desID)
		r := call(t, "PUT", base, token, map[string]interface{}{
			"schedules": []map[string]interface{}{
				{"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00"},
				{"day_of_week": 1, "departure_time": "08:00", "remind_time": "21:00"},
			},
		})
		if r.Status == 200 {
			t.Error("hari dobel di payload = 200, seharusnya ditolak")
		}
		if countRows(t, "destination_schedules", "destination_id = ?", desID) != before {
			t.Error("ada baris tersimpan padahal payload tidak valid")
		}
	})

	t.Run("mengganti jadwal menghapus reminder lamanya", func(t *testing.T) {
		makeAllDaysSchedule(t, token, desID)
		generateReminders(t)
		if countRows(t, "reminders", "destination_id = ?", desID) == 0 {
			t.Fatal("tidak ada reminder ter-generate")
		}

		// Ganti jadi 1 hari saja -> reminder lama harus hilang (cascade FK).
		r := call(t, "PUT", base, token, map[string]interface{}{
			"schedules": []map[string]interface{}{
				{"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00"},
			},
		})
		wantStatus(t, r, 200, "replace 1 hari")
		if countRows(t, "reminders", "destination_id = ?", desID) != 0 {
			t.Error("reminder lama masih ada, padahal jadwalnya sudah diganti (cascade gagal)")
		}
	})
}

func TestScheduleListOrderAndDetail(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/schedules")

	// Insert dalam urutan acak supaya urutan list ikut teruji.
	for _, day := range []int{5, 1, 6, 0, 3} {
		call(t, "POST", base, token, map[string]interface{}{
			"day_of_week": day, "departure_time": "07:00", "remind_time": "20:00",
		})
	}

	r := call(t, "GET", base, token, nil)
	wantStatus(t, r, 200, "list schedule")
	list := r.list()
	if len(list) != 5 {
		t.Fatalf("jumlah = %d, want 5", len(list))
	}
	prev := -1.0
	for _, row := range list {
		m, _ := row.(map[string]interface{})
		day := m["day_of_week"].(float64)
		if day < prev {
			t.Errorf("list tidak urut naik: %v setelah %v", day, prev)
		}
		prev = day
	}

	// Detail juga harus mengembalikan day_of_week yang benar.
	firstRow, _ := list[0].(map[string]interface{})
	firstID := uint64(firstRow["id"].(float64))
	r = call(t, "GET", base+"/"+u64s(firstID), token, nil)
	wantStatus(t, r, 200, "detail schedule")
	if got := r.data()["day_of_week"]; got != float64(0) {
		t.Errorf("detail day_of_week = %v, want 0 (Minggu)", got)
	}
}
