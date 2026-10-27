# Test Integrasi

Test ini menembak HTTP layer sungguhan (middleware -> handler -> service -> GORM -> MySQL)
lewat `app.Test()`, jadi hampir semua bug jalur database ikut tertangkap.

## Prasyarat

1. MySQL jalan.
2. Ada database terpisah untuk test (JANGAN pakai database dev).
3. Migrasi sudah dijalankan di database tersebut.

```bash
# 1. buat database test
mysql -h 127.0.0.1 -P 3306 -u root -p -e \
  "CREATE DATABASE byl_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. jalankan migrasi ke database test
migrate -database "mysql://root:PASSWORD@tcp(127.0.0.1:3306)/byl_test" \
  -path databases/migrations up
```

Kredensial ada di `tests/harness_test.go` bagian `TestMain`. `godotenv.Load()` tidak
menimpa variabel env yang sudah di-set, jadi nilai di situ yang dipakai.

## Menjalankan

```bash
go test ./tests/ -v
go test ./tests/ -count=1          # tanpa cache
go test ./tests/ -race             # race detector
go test ./tests/ -run TestReminder # subset
```

## Yang diuji

| File | Cakupan |
|---|---|
| `auth_test.go` | register, login, `/me`, penolakan token, isolasi antar user |
| `destination_test.go` | CRUD, preload jadwal/item/reminder, filter, pagination |
| `checklist_item_test.go` | CRUD, nama unik per destinasi, cascade item check |
| `schedule_test.go` | CRUD, validasi hari/jam, urutan jam hari-H, bulk replace (rollback) |
| `default_tag_test.go` | regresi tag `default:` GORM, libur total & per-hari |
| `reminder_worker_test.go` | generate, idempotensi, backfill item, dispatch pending->sent |
| `reminder_api_test.go` | list/filter/pagination, detail+progress, read, skip, centang item |
| `integrity_test.go` | cascade delete, deteksi baris yatim |

Setiap test memanggil `resetDB` yang me-`TRUNCATE` semua tabel (dengan
`FOREIGN_KEY_CHECKS=0`), jadi test bisa dijalankan berulang dan urutan tidak berpengaruh.

## Catatan

- `REMINDER_HORIZON_DAYS` di-`TestMain` = 14. Jadwal aktif setiap hari dengan H-1 23:00
  menghasilkan tepat 14 reminder (kemunculan hari ini selalu ter-skip karena
  `remind_at`-nya sudah lewat), jadi jumlahnya bisa diprediksi tanpa bergantung tanggal.
- Data dev tidak pernah disentuh: test hanya menulis ke `byl_test`.