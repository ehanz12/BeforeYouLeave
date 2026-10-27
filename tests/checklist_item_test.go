package tests

import "testing"

func TestChecklistItemCRUD(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/items")

	// Create
	r := call(t, "POST", base, token, map[string]interface{}{"name": "Buku Tulis", "is_required": true})
	wantStatus(t, r, 200, "create item")
	id := idOf(t, r, "create item")
	if got := r.data()["destination_id"]; uint64(got.(float64)) != desID {
		t.Errorf("destination_id = %v, want %d", got, desID)
	}
	if r.data()["is_required"] != true {
		t.Error("is_required = false, want true")
	}

	// Detail
	r = call(t, "GET", base+"/"+u64s(id), token, nil)
	wantStatus(t, r, 200, "detail item")
	if r.data()["name"] != "Buku Tulis" {
		t.Errorf("name = %v, want Buku Tulis", r.data()["name"])
	}

	// Update sebagian: hanya is_required
	r = call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{"is_required": false})
	wantStatus(t, r, 200, "patch is_required")
	if r.data()["is_required"] != false {
		t.Error("is_required tidak jadi false")
	}
	if r.data()["name"] != "Buku Tulis" {
		t.Errorf("name = %v, harusnya tetap Buku Tulis", r.data()["name"])
	}

	// Update name
	r = call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{"name": "Buku Tulis Hijau"})
	wantStatus(t, r, 200, "patch name")
	if r.data()["name"] != "Buku Tulis Hijau" {
		t.Errorf("name = %v", r.data()["name"])
	}

	// Delete
	r = call(t, "DELETE", base+"/"+u64s(id), token, nil)
	wantStatus(t, r, 200, "delete item")
	if countRows(t, "checklist_items", "id = ?", id) != 0 {
		t.Error("item masih ada setelah delete")
	}
}

func TestChecklistItemValidation(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/items")

	t.Run("nama kosong ditolak", func(t *testing.T) {
		r := call(t, "POST", base, token, map[string]string{"name": ""})
		if r.Status == 200 {
			t.Error("nama kosong = 200, seharusnya ditolak")
		}
	})

	t.Run("nama duplikat dalam 1 destinasi ditolak", func(t *testing.T) {
		makeItem(t, token, desID, "Buku")
		r := call(t, "POST", base, token, map[string]string{"name": "Buku"})
		if r.Status == 200 {
			t.Error("nama duplikat = 200, seharusnya ditolak")
		}
		if countRows(t, "checklist_items", "destination_id = ? AND name = ?", desID, "Buku") != 1 {
			t.Error("harusnya tetap 1 baris, ditemukan duplikat di DB")
		}
	})

	t.Run("nama sama di demisekai destinasi boleh", func(t *testing.T) {
		other := makeDestination(t, token, "Kampus")
		r := call(t, "POST", nested("/api/destination", other, "/items"), token, map[string]string{"name": "Buku"})
		wantStatus(t, r, 200, "nama sama di destinasi lain")
	})

	t.Run("update ke nama sendiri tidak dianggap duplikat", func(t *testing.T) {
		id := firstItemID(t, desID)
		r := call(t, "PATCH", base+"/"+u64s(id), token, map[string]interface{}{"is_required": true})
		wantStatus(t, r, 200, "update tanpa ubah nama")
	})

	t.Run("update ke nama milik item lain ditolak", func(t *testing.T) {
		makeItem(t, token, desID, "Penghapus")
		makeItem(t, token, desID, "Pensil")
		ids := itemIDs(t, desID)
		r := call(t, "PATCH", base+"/"+u64s(ids[0]), token, map[string]string{"name": "Pensil"})
		if r.Status == 200 {
			t.Error("update ke nama item lain = 200, seharusnya ditolak")
		}
	})

	t.Run("patch tanpa field ditolak", func(t *testing.T) {
		id := firstItemID(t, desID)
		r := call(t, "PATCH", base+"/"+u64s(id), token, map[string]string{})
		if r.Status == 200 {
			t.Error("patch kosong = 200, seharusnya ditolak")
		}
	})

	t.Run("destination tidak ada", func(t *testing.T) {
		r := call(t, "POST", "/api/destination/999999/items", token, map[string]string{"name": "X"})
		wantStatus(t, r, 400, "destinasi tidak ada")
	})

	t.Run("item tidak ada", func(t *testing.T) {
		r := call(t, "GET", base+"/999999", token, nil)
		wantStatus(t, r, 404, "item tidak ada")
	})
}

func TestChecklistItemList(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")
	base := nested("/api/destination", desID, "/items")

	makeItem(t, token, desID, "Buku Tulis")
	makeItem(t, token, desID, "Pena")
	makeItem(t, token, desID, "Botol Minum")

	t.Run("semua item", func(t *testing.T) {
		r := call(t, "GET", base, token, nil)
		wantStatus(t, r, 200, "list")
		if n := len(r.list()); n != 3 {
			t.Errorf("jumlah item = %d, want 3", n)
		}
	})

	t.Run("item tidak ikut dari destination lain", func(t *testing.T) {
		// Item harus ter-scope ke destination ini saja.
		other := makeDestination(t, token, "Kampus")
		makeItem(t, token, other, "Laptop")
		r := call(t, "GET", base, token, nil)
		if n := len(r.list()); n != 3 {
			t.Errorf("item dari destination lain bocor: total = %d, want 3", n)
		}
	})

	t.Run("filter search + is_required", func(t *testing.T) {
		r := call(t, "GET", base+"?search=pen", token, nil)
		wantStatus(t, r, 200, "search")
		if n := len(r.list()); n != 1 {
			t.Errorf("search 'pen' = %d, want 1", n)
		}
		r = call(t, "GET", base+"?is_required=true", token, nil)
		wantStatus(t, r, 200, "is_required")
		if n := len(r.list()); n != 3 {
			t.Errorf("is_required=true = %d, want 3 (semua dibuat required)", n)
		}
	})

	t.Run("item muncul di detail destinasi", func(t *testing.T) {
		r := call(t, "GET", "/api/destination/"+u64s(desID), token, nil)
		wantStatus(t, r, 200, "detail destination")
		items, _ := r.data()["checklist_items"].([]interface{})
		if len(items) != 3 {
			t.Errorf("checklist_items di detail = %d, want 3", len(items))
		}
	})
}

// Hapus item harus ikut menyapu reminder yang sudah mereferensikannya.
func TestChecklistItemDeleteCascadesItemCheck(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")
	desID := makeDestination(t, token, "Sekolah")

	item := makeItem(t, token, desID, "Buku")
	makeAllDaysSchedule(t, token, desID)
	generateReminders(t)

	if countRows(t, "reminder_item_checks", "checklist_item_id = ?", item) == 0 {
		t.Fatal("tidak ada reminder_item_checks, test tidak bermakna")
	}

	r := call(t, "DELETE", nested("/api/destination", desID, "/items")+"/"+u64s(item), token, nil)
	wantStatus(t, r, 200, "delete item")

	if countRows(t, "reminder_item_checks", "checklist_item_id = ?", item) != 0 {
		t.Error("reminder_item_checks tidak ikut terhapus (cascade FK gagal)")
	}
}

// ---------------------------------------------------------------- helpers

func firstItemID(t *testing.T, destinationID uint64) uint64 {
	t.Helper()
	ids := itemIDs(t, destinationID)
	if len(ids) == 0 {
		t.Fatal("tidak ada item")
	}
	return ids[0]
}

func itemIDs(t *testing.T, destinationID uint64) []uint64 {
	t.Helper()
	var ids []uint64
	err := db().Table("checklist_items").
		Where("destination_id = ?", destinationID).
		Order("id ASC").Pluck("id", &ids).Error
	if err != nil {
		t.Fatal(err)
	}
	return ids
}
