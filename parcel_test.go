package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
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
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	ind, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, ind)
	parcel.Number = ind

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	result, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parcel, result)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	require.NoError(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	ind, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, ind)
	parcel.Number = ind

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	parcelGet, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.NotEqual(t, parcel, parcelGet)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	parcel.Client = 1001
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	ind, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, ind)
	parcel.Number = ind
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	client, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.NotEqual(t, parcel.Status, client.Status)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
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
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем в карту
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		assert.True(t, ok, "unexpected parcel with number: %d", parcel.Number)
		assert.Equal(t, expected, parcel)
	}
}
