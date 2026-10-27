package tests

import (
	"testing"
	"time"
)

func TestReminderList(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeItem(t, token, desID, "Buku")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	// Total reminder = 14 (horizon 14, hari ini ter-skip).
	const total = 14

	t.Run("default pagination dibatasi 10", func(t *testing.T) {
		// utils.NormalizePagination punya defaultLimit=10, jadi tanpa ?limit
		// hanya 10 yang dikembalikan meski totalnya 14.
		r := call(t, "GET", "/api/reminder", token, nil)
		wantStatus(t, r, 200, "list")
		if n := len(r.list()); n != 10 {
			t.Errorf("jumlah = %d, want 10 (default limit)", n)
		}
		meta, _ := r.Body["meta"].(map[string]interface{})
		if got := int(meta["total"].(float64)); got != total {
			t.Errorf("meta.total = %d, want %d", got, total)
		}
	})

	t.Run("limit eksplisit mengambil semua", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?limit=100", token, nil)
		wantStatus(t, r, 200, "list")
		if n := len(r.list()); n != total {
			t.Errorf("jumlah = %d, want %d", n, total)
		}
	})

	t.Run("urut dari yang paling dekat", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?limit=100", token, nil)
		prev := ""
		for _, row := range r.list() {
			m := row.(map[string]interface{})
			cur := m["remind_at"].(string)
			if prev != "" && cur < prev {
				t.Errorf("tidak urut: %s muncul setelah %s", cur, prev)
			}
			prev = cur
		}
	})

	t.Run("punya destination_name", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?limit=100", token, nil)
		row := r.list()[0].(map[string]interface{})
		if row["destination_name"] != "Sekolah" {
			t.Errorf("destination_name = %v, want Sekolah", row["destination_name"])
		}
		if row["destination_id"] == nil {
			t.Error("destination_id null")
		}
	})

	t.Run("filter status", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?status=pending&limit=100", token, nil)
		wantStatus(t, r, 200, "status pending")
		if n := len(r.list()); n != total {
			t.Errorf("pending = %d, want %d", n, total)
		}
		r = call(t, "GET", "/api/reminder?status=read&limit=100", token, nil)
		if n := len(r.list()); n != 0 {
			t.Errorf("read = %d, want 0", n)
		}
	})

	t.Run("status ngawur ditolak", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?status=bogus", token, nil)
		if r.Status == 200 {
			t.Error("status ngawur = 200, seharusnya ditolak")
		}
	})

	t.Run("unread_only", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?unread_only=true&limit=100", token, nil)
		wantStatus(t, r, 200, "unread_only")
		if n := len(r.list()); n != total {
			t.Errorf("unread_only = %d, want %d (semua pending)", n, total)
		}

		// Tandai satu jadi read, harus berkurang.
		first := uint64(r.list()[0].(map[string]interface{})["id"].(float64))
		call(t, "PATCH", "/api/reminder/"+u64s(first)+"/read", token, nil)

		r = call(t, "GET", "/api/reminder?unread_only=true&limit=100", token, nil)
		if n := len(r.list()); n != total-1 {
			t.Errorf("unread_only setelah 1 di-read = %d, want %d", n, total-1)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder?page=1&limit=5", token, nil)
		wantStatus(t, r, 200, "page 1")
		if n := len(r.list()); n != 5 {
			t.Errorf("halaman 1 = %d, want 5", n)
		}
		meta, _ := r.Body["meta"].(map[string]interface{})
		if got := int(meta["total"].(float64)); got != total {
			t.Errorf("meta.total = %d, want %d", got, total)
		}
		if got := int(meta["total_pages"].(float64)); got != 3 {
			t.Errorf("meta.total_pages = %d, want 3", got)
		}

		// Halaman terakhir tidak boleh mengulang baris halaman 1.
		firstPage := call(t, "GET", "/api/reminder?page=1&limit=5", token, nil)
		lastPage := call(t, "GET", "/api/reminder?page=3&limit=5", token, nil)
		seen := map[float64]bool{}
		for _, row := range firstPage.list() {
			seen[row.(map[string]interface{})["id"].(float64)] = true
		}
		for _, row := range lastPage.list() {
			id := row.(map[string]interface{})["id"].(float64)
			if seen[id] {
				t.Errorf("id %v muncul di halaman 1 dan 3 (duplikat)", id)
			}
		}
		if n := len(lastPage.list()); n != 4 {
			t.Errorf("halaman 3 = %d, want 4", n)
		}
	})

	t.Run("filter tanggal", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		r := call(t, "GET", "/api/reminder?start_date="+today+"&limit=100", token, nil)
		wantStatus(t, r, 200, "start_date")
		if n := len(r.list()); n == 0 {
			t.Error("start_date hari ini menghasilkan 0 reminder")
		}
		r = call(t, "GET", "/api/reminder?start_date=bukan-tanggal", token, nil)
		if r.Status == 200 {
			t.Error("start_date ngawur = 200, seharusnya ditolak")
		}
	})
}

// Dua destination harus bisa difilter terpisah tanpa saling bocor.
func TestReminderListFilterByDestination(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	sekolah := makeDestination(t, token, "Sekolah")
	kantor := makeDestination(t, token, "Kantor")

	// Sekolah aktif tiap hari (H-1 23:00).
	makeAllDaysSchedule(t, token, sekolah)

	// Kantor hanya Jumat + Sabtu, dan mode hari-H.
	r := call(t, "PUT", nested("/api/destination", kantor, "/schedules"), token, map[string]interface{}{
		"schedules": []map[string]interface{}{
			{"day_of_week": 5, "departure_time": "18:00", "remind_time": "17:00", "remind_d_minus": 0},
			{"day_of_week": 6, "departure_time": "09:00", "remind_time": "08:00", "remind_d_minus": 0},
		},
	})
	wantStatus(t, r, 200, "jadwal kantor")

	generateReminders(t)

	r = call(t, "GET", "/api/reminder?destination_id="+u64s(sekolah)+"&limit=100", token, nil)
	wantStatus(t, r, 200, "filter sekolah")
	rows := r.list()
	sekolahRows := rows
	if len(rows) != 14 {
		t.Errorf("reminder sekolah = %d, want 14", len(rows))
	}
	for _, row := range rows {
		if got := row.(map[string]interface{})["destination_name"]; got != "Sekolah" {
			t.Errorf("destination_name = %v, want Sekolah", got)
		}
	}

	r = call(t, "GET", "/api/reminder?destination_id="+u64s(kantor)+"&limit=100", token, nil)
	wantStatus(t, r, 200, "filter kantor")
	rows = r.list()
	// Hanya 2 hari dalam 14 hari horizon, jadi jumlahnya kecil. Yang penting:
	// tidak boleh ada reminder sekolah yang bocor ke filter ini.
	for _, row := range rows {
		m := row.(map[string]interface{})
		if m["destination_name"] != "Kantor" {
			t.Errorf("destination_name = %v, want Kantor", m["destination_name"])
		}
	}
	// Horizon 14 hari = 2 minggu penuh, jadi Jumat dan Sabtu masing-masing muncul
	// 2 kali. Kalau kemunculan hari ini sudah lewat waktunya (misal hari ini
	// Jumat dan remind 17:00 sementara sekarang sudah malam), satu lagi hilang.
	if len(rows) < 3 || len(rows) > 4 {
		t.Errorf("reminder kantor = %d, want 3 atau 4 (2x Jumat + 2x Sabtu)", len(rows))
	}
	if len(rows) >= len(sekolahRows) {
		t.Errorf("kantor (%d) tidak boleh sebanyak sekolah (%d)", len(rows), len(sekolahRows))
	}
}

func TestReminderDetailAndProgress(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	buku := makeItem(t, token, desID, "Buku")
	pena := makeItem(t, token, desID, "Pena")
	botol := makeItem(t, token, desID, "Botol")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	id := firstReminderID(t, desID)

	r := call(t, "GET", "/api/reminder/"+u64s(id), token, nil)
	wantStatus(t, r, 200, "detail reminder")
	data := r.data()

	checks, _ := data["item_checks"].([]interface{})
	if len(checks) != 3 {
		t.Fatalf("item_checks = %d, want 3", len(checks))
	}
	// Setiap item harus ikut membawa detail barangnya.
	first, _ := checks[0].(map[string]interface{})
	item, _ := first["checklist_item"].(map[string]interface{})
	if item == nil || item["name"] == nil {
		t.Error("checklist_item tidak ikut di-preload")
	}

	progress, _ := data["progress"].(map[string]interface{})
	if got := int(progress["total"].(float64)); got != 3 {
		t.Errorf("progress.total = %d, want 3", got)
	}
	if got := int(progress["checked"].(float64)); got != 0 {
		t.Errorf("progress.checked = %d, want 0", got)
	}

	dest, _ := data["destination"].(map[string]interface{})
	if dest == nil || dest["name"] != "Sekolah" {
		t.Error("destination tidak ikut di-preload")
	}

	// Centang satu item.
	r = call(t, "PATCH", "/api/reminder/"+u64s(id)+"/items/"+u64s(buku), token,
		map[string]interface{}{"is_checked": true})
	wantStatus(t, r, 200, "centang")
	progress, _ = r.data()["progress"].(map[string]interface{})
	if got := int(progress["checked"].(float64)); got != 1 {
		t.Errorf("setelah centang 1, progress.checked = %d, want 1", got)
	}
	if got := int(progress["total"].(float64)); got != 3 {
		t.Errorf("progress.total berubah jadi %d, want 3", got)
	}
	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND is_checked = 1", id, buku); got != 1 {
		t.Error("centangan tidak tersimpan")
	}
	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND checked_at IS NOT NULL", id, buku); got != 1 {
		t.Error("checked_at tidak diisi saat dicentang")
	}

	// Centang item kedua.
	call(t, "PATCH", "/api/reminder/"+u64s(id)+"/items/"+u64s(pena), token,
		map[string]interface{}{"is_checked": true})
	// Uncentang item pertama.
	r = call(t, "PATCH", "/api/reminder/"+u64s(id)+"/items/"+u64s(buku), token,
		map[string]interface{}{"is_checked": false})
	wantStatus(t, r, 200, "uncentang")
	progress, _ = r.data()["progress"].(map[string]interface{})
	if got := int(progress["checked"].(float64)); got != 1 {
		t.Errorf("setelah uncentang, checked = %d, want 1", got)
	}
	// checked_at harus dikosongkan lagi saat di-uncheck.
	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND checked_at IS NOT NULL", id, buku); got != 0 {
		t.Error("checked_at tidak dikosongkan saat di-uncheck")
	}

	// Centangan hanya untuk reminder itu, tidak boleh bocor ke reminder lain.
	other := secondReminderID(t, desID, id)
	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND is_checked = 1", other, pena); got != 0 {
		t.Error("centangan bocor ke reminder lain")
	}

	_ = botol
	_ = item
}

func TestReminderReadAndSkip(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	id := firstReminderID(t, desID)

	t.Run("tandai read", func(t *testing.T) {
		r := call(t, "PATCH", "/api/reminder/"+u64s(id)+"/read", token, nil)
		wantStatus(t, r, 200, "read")
		if r.data()["status"] != "read" {
			t.Errorf("status = %v, want read", r.data()["status"])
		}
		if r.data()["read_at"] == nil {
			t.Error("read_at tidak diisi")
		}
	})

	t.Run("read bersifat idempotent", func(t *testing.T) {
		var firstRead time.Time
		db().Table("reminders").Where("id = ?", id).Select("read_at").Scan(&firstRead)

		time.Sleep(1100 * time.Millisecond)
		r := call(t, "PATCH", "/api/reminder/"+u64s(id)+"/read", token, nil)
		wantStatus(t, r, 200, "read kedua")

		var secondRead time.Time
		db().Table("reminders").Where("id = ?", id).Select("read_at").Scan(&secondRead)
		if !secondRead.Equal(firstRead) {
			t.Error("read_at bergeser saat read dipanggil ulang, harusnya tetap sama")
		}
	})

	t.Run("skip", func(t *testing.T) {
		other := secondReminderID(t, desID, id)
		r := call(t, "PATCH", "/api/reminder/"+u64s(other)+"/skip", token, nil)
		wantStatus(t, r, 200, "skip")
		if r.data()["status"] != "skipped" {
			t.Errorf("status = %v, want skipped", r.data()["status"])
		}
		// Skip tidak boleh mengubah jadwalnya.
		if countRows(t, "destination_schedules", "destination_id = ?", desID) != 7 {
			t.Error("skip reminder ikut menghapus jadwal")
		}
	})

	t.Run("reminder tidak ada", func(t *testing.T) {
		r := call(t, "GET", "/api/reminder/999999", token, nil)
		wantStatus(t, r, 404, "tidak ada")
		r = call(t, "PATCH", "/api/reminder/999999/read", token, nil)
		wantStatus(t, r, 404, "read tidak ada")
	})
}

// Item dari destination lain tidak boleh bisa dicentang ke reminder ini.
func TestReminderItemCheckRejectsForeignItem(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	desA := makeDestination(t, token, "Sekolah")
	makeItem(t, token, desA, "Buku")
	makeAllDaysSchedule(t, token, desA)
	generateReminders(t)

	desB := makeDestination(t, token, "Kantor")
	foreign := makeItem(t, token, desB, "Laptop")

	reminder := firstReminderID(t, desA)
	r := call(t, "PATCH", "/api/reminder/"+u64s(reminder)+"/items/"+u64s(foreign), token,
		map[string]interface{}{"is_checked": true})
	if r.Status == 200 {
		t.Error("item milik destination lain bisa dicentang, seharusnya ditolak")
	}
	if got := countRows(t, "reminder_item_checks", "checklist_item_id = ?", foreign); got != 0 {
		t.Error("baris reminder_item_checks tetap dibuat padahal item bukan miliknya")
	}
}

// Barang yang ditambahkan SETELAH reminder dibuat harus tetap bisa dicentang
// (barisnya dibuat on-demand).
func TestReminderItemCheckOnLateAddedItem(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	reminder := firstReminderID(t, desID)
	if got := countRows(t, "reminder_item_checks", "reminder_id = ?", reminder); got != 0 {
		t.Fatalf("reminder awal punya %d item check, test tidak bermakna", got)
	}

	late := makeItem(t, token, desID, "Barang Terlambat")

	r := call(t, "PATCH", "/api/reminder/"+u64s(reminder)+"/items/"+u64s(late), token,
		map[string]interface{}{"is_checked": true})
	wantStatus(t, r, 200, "centang barang baru")
	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND is_checked = 1", reminder, late); got != 1 {
		t.Error("barang yang ditambahkan belakangan tidak bisa dicentang")
	}
}

// Menandai read tidak boleh mengubah centangan barang.
func TestMarkReadKeepsItemChecks(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	item := makeItem(t, token, desID, "Buku")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	id := firstReminderID(t, desID)
	call(t, "PATCH", "/api/reminder/"+u64s(id)+"/items/"+u64s(item), token,
		map[string]interface{}{"is_checked": true})
	call(t, "PATCH", "/api/reminder/"+u64s(id)+"/read", token, nil)

	if got := countRows(t, "reminder_item_checks",
		"reminder_id = ? AND checklist_item_id = ? AND is_checked = 1", id, item); got != 1 {
		t.Error("menandai read menghapus centangan barang")
	}
}

func secondReminderID(t *testing.T, destinationID, exclude uint64) uint64 {
	t.Helper()
	var id uint64
	err := db().Table("reminders").
		Where("destination_id = ? AND id != ?", destinationID, exclude).
		Order("remind_at ASC").Limit(1).Select("id").Scan(&id).Error
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("tidak ada reminder kedua")
	}
	return id
}
