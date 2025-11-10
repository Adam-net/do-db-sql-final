package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
	//"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}
	// get

	p, err := store.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	parcel.Number = id
	if parcel != p {
		t.Fatalf("Parcel does not match expected parcel")
	}

	// delete

	err = store.Delete(id)
	if err != nil {
		t.Fatal(err)
	}

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}

	// set address

	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatal(err)
	}

	var res string

	row := db.QueryRow("SELECT address FROM parcel WHERE number = ?", id)
	err = row.Scan(&res)
	if err != nil {
		t.Fatal(err)
	}
	if res != newAddress {
		t.Fatalf("Address does not match expected address")
	}
	// check

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}

	// set status

	newStatus := "new status"
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatal(err)
	}

	var res string

	row := db.QueryRow("SELECT status FROM parcel WHERE number = ?", id)
	err = row.Scan(&res)
	if err != nil {
		t.Fatal(err)
	}
	if res != newStatus {
		t.Fatalf("Address does not match expected address")
	}

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		res, err := db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
			parcels[i].Client, parcels[i].Status, parcels[i].Address, parcels[i].CreatedAt)
		if err != nil {
			t.Fatal(err)
		}

		id_temp, err := res.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		id := int(id_temp)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	rows, err := db.Query("SELECT number FROM parcel WHERE client = ?", client)
	if err != nil {
		t.Fatal(err)
	}
	var storedParcels []int64
	var temp int64

	count := 0
	for rows.Next() {
		count++
		err := rows.Scan(&temp)
		if err != nil {
			t.Fatal(err)
		}
		storedParcels = append(storedParcels, temp)
	}
	if count != len(parcels) {
		t.Fatalf("Number of parcels does not match expected parcel count")
	}

	// check
	for _, parcel := range storedParcels {

		if _, ok := parcelMap[int(parcel)]; !ok {
			t.Fatalf("Parcel does not match expected parcel")
		}
	}
}
