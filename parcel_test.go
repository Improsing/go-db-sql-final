package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	defer db.Close()
	
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в бд
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)

	// get
	// получаем только что добавленную посылку, убеждаемся в отсутствии ошибки
	// проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	gotParcel, err := store.Get(parcelID)
	require.NoError(t, err)
	gotParcel.Number = parcel.Number
	assert.Equal(t, parcel, gotParcel)

	// delete
	// удаляем добавленную посылку, убеждаемся в отсутствии ошибки
	// проверяем, что посылку больше нельзя получить из БД
	err = store.Delete(parcelID)
	require.NoError(t, err)

	got, err := store.Get(parcelID)
	require.Equal(t, sql.ErrNoRows, err)
	require.Empty(t, got)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)
	// set address
	// обновляем адрес, убеждаемся в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcelID, newAddress)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и убеждаемся, что адрес обновился
	got, err := store.Get(parcelID)
	require.NoError(t, err)
	assert.Equal(t, newAddress, got.Address)  
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusDelivered

	err = store.SetStatus(parcel.Number, newStatus)

	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	storedParcel, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	// if err != nil {
	require.NoError(t, err)
	// }
	defer db.Close()

	store := NewParcelStore(db)
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
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])

		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор у добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		p, ok := parcelMap[parcel.Number]
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		require.True(t, ok)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, p, parcel)
	}
}