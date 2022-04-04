package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(2)
		c.Set("key1", 1)
		c.Set("key2", 2)
		c.Set("key3", 3)
		//key3, key2
		val, ok := c.Get("key1") // проверка выталкивания по очереди добавления
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("key2") // проверка выталкивания по очереди запроса
		// key2, key3
		c.Set("key4", 4)
		//key4, key2
		val, ok = c.Get("key3")
		require.False(t, ok)
		require.Nil(t, val)

		c.Set("key2", 22) // проверка обновления значения и выталкивания по очереди обновлнеия
		//key2, key4
		val, ok = c.Get("key2")
		require.True(t, ok)
		require.Equal(t, val, 22)
		c.Set("key4", 44)
		c.Set("key5", 5)
		//key5, key4
		val, ok = c.Get("key2")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("Clear", func(t *testing.T) {
		c := NewCache(5)
		c.Set("key1", 1)
		c.Set("key2", 2)
		c.Set("key3", 3)
		//key3, key2, key1

		c.Clear()
		val, ok := c.Get("key1")
		require.False(t, ok)
		require.Nil(t, val)
		val, ok = c.Get("key2")
		require.False(t, ok)
		require.Nil(t, val)
		val, ok = c.Get("key3")
		require.False(t, ok)
		require.Nil(t, val)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
