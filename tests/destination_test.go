package tests

import "testing"

func TestDestinationCRUD(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	// Create
	r := call(t, "POST", "/api/destination", token, map[string]interface{}{
		"name": "Sekolah", "description": "SD Negeri 1",
		"latitude": -6.2, "longitude": 106.8, "radius": 5.5,
	})
	wantStatus(t, r, 200, "create")
	id := idOf(t, r, "create")

	if got := r.data()["name"]; got != "Sekolah" {
		t.Errorf("name = %v, want Sekolah", got)
	}
	// nilai koordinat harus utuh, tidak berubah jadi 0 / null
	if got := r.data()["radius"]; got != 5.5 {
		t.Errorf("radius = %v, want 5.5", got)
	}

	// Detail harus preload jadwal + checklist (meski kosong, harus array bukan null)
	r = call(t, "GET", "/api/destination/"+u64s(id), token, nil)
	wantStatus(t, r, 200, "detail")
	for _, key := range []string{"schedules", "checklist_items", "reminders"} {
		v, ok := r.data()[key]
		if !ok {
			t.Errorf("detail tidak punya field %q", key)
			continue
		}
		if v == nil {
			t.Errorf("field %q = null, FE akan dapat null bukan array", key)
		}
	}

	// Update sebagian (PATCH, hanya name)
	r = call(t, "PATCH", "/api/destination/"+u64s(id), token, map[string]string{"name": "Sekolah Baru"})
	wantStatus(t, r, 200, "update")
	if got := r.data()["name"]; got != "Sekolah Baru" {
		t.Errorf("name = %v, want Sekolah Baru", got)
	}
	// field lain harus tetap
	if got := r.data()["description"]; got != "SD Negeri 1" {
		t.Errorf("description = %v, harusnya tetap SD Negeri 1", got)
	}

	// PATCH tanpa field apa pun harus ditolak
	r = call(t, "PATCH", "/api/destination/"+u64s(id), token, map[string]string{})
	if r.Status == 200 {
		t.Error("PATCH tanpa field = 200, seharusnya ditolak")
	}

	// Delete
	r = call(t, "DELETE", "/api/destination/"+u64s(id), token, nil)
	wantStatus(t, r, 200, "delete")
	if countRows(t, "destinations", "id = ?", id) != 0 {
		t.Error("destinasi masih ada setelah delete")
	}

	// Detail yang sudah dihapus harus 404
	r = call(t, "GET", "/api/destination/"+u64s(id), token, nil)
	wantStatus(t, r, 404, "detail setelah delete")
}

func TestDestinationValidation(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	t.Run("nama kosong ditolak", func(t *testing.T) {
		r := call(t, "POST", "/api/destination", token, map[string]string{"name": ""})
		if r.Status == 200 {
			t.Error("nama kosong = 200, seharusnya ditolak")
		}
	})

	t.Run("nama duplikat ditolak", func(t *testing.T) {
		makeDestination(t, token, "Kampus")
		r := call(t, "POST", "/api/destination", token, map[string]string{"name": "Kampus"})
		wantStatus(t, r, 500, "nama duplikat (handler balas 500)")
		if r.errMsg() != "nama sudah dipakai" {
			t.Errorf("pesan = %q, want 'nama sudah dipakai'", r.errMsg())
		}
	})

	t.Run("nama sama di user lain boleh", func(t *testing.T) {
		makeUser(t, "B", "b@test.com")
		tokenB := tokenFor(t, "b@test.com")
		makeDestination(t, tokenB, "Kampus")
		if countRows(t, "destinations", "name = ?", "Kampus") != 2 {
			t.Error("nama_board uniqueness seharusnya per-user, tapi hanya 1 baris")
		}
	})

	t.Run("id bukan angka", func(t *testing.T) {
		r := call(t, "GET", "/api/destination/abc", token, nil)
		wantStatus(t, r, 400, "id bukan angka")
	})
}

func TestDestinationListFilterAndPagination(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	for _, name := range []string{"Sekolah", "Kampus", "Kantor"} {
		makeDestination(t, token, name)
	}

	t.Run("filter search", func(t *testing.T) {
		r := call(t, "GET", "/api/destination?search=Kam", token, nil)
		wantStatus(t, r, 200, "search")
		if n := len(r.list()); n != 1 {
			t.Errorf("search 'Kam' = %d hasil, want 1", n)
		}
	})

	t.Run("filter is_active", func(t *testing.T) {
		r := call(t, "GET", "/api/destination?is_active=true", token, nil)
		wantStatus(t, r, 200, "is_active")
		if n := len(r.list()); n != 3 {
			t.Errorf("is_active=true = %d, want 3", n)
		}
		r = call(t, "GET", "/api/destination?is_active=false", token, nil)
		wantStatus(t, r, 200, "is_active=false")
		if n := len(r.list()); n != 0 {
			t.Errorf("is_active=false = %d, want 0", n)
		}
	})

	t.Run("pagination limit dibatasi", func(t *testing.T) {
		// utils.NormalizePagination punya maxLimit=100
		r := call(t, "GET", "/api/destination?page=1&limit=9999", token, nil)
		wantStatus(t, r, 200, "limit besar")
		meta, _ := r.Body["meta"].(map[string]interface{})
		if got := int(meta["limit"].(float64)); got != 100 {
			t.Errorf("limit = %d, want 100 (max)", got)
		}
	})

	t.Run("page di luar jangkauan", func(t *testing.T) {
		r := call(t, "GET", "/api/destination?page=99&limit=10", token, nil)
		wantStatus(t, r, 200, "page 99")
		if n := len(r.list()); n != 0 {
			t.Errorf("page 99 = %d hasil, want 0", n)
		}
	})

	t.Run("filter tanggal tidak valid", func(t *testing.T) {
		r := call(t, "GET", "/api/destination?start_date=bukan-tanggal", token, nil)
		if r.Status == 200 {
			t.Error("start_date ngawur = 200, seharusnya 400")
		}
	})
}
