package recached

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test with a simple string value
	updateCount := 0
	updateFunc := func() (string, error) {
		updateCount++
		return "test value", nil
	}

	// Create a new cache
	cache := New(ctx, 100*time.Millisecond, updateFunc)

	// Verify that the cache was initialized with the correct value
	if got := cache.Get(); got != "test value" {
		t.Errorf("Initial value = %v, want %v", got, "test value")
	}

	// Verify that updateFunc was called once during initialization
	if updateCount != 1 {
		t.Errorf("updateCount = %v, want %v", updateCount, 1)
	}
}

func TestGet(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a cache with an integer value
	initialValue := 42
	updateFunc := func() (int, error) {
		return initialValue, nil
	}

	cache := New(ctx, time.Hour, updateFunc)

	// Verify that Get returns the cached value
	if got := cache.Get(); got != initialValue {
		t.Errorf("Get() = %v, want %v", got, initialValue)
	}
}

func TestUpdate(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a value that will change when Update is called
	value := 1
	updateFunc := func() (int, error) {
		value++
		return value, nil
	}

	cache := New(ctx, time.Hour, updateFunc)

	// Initial value should be 2 (1 incremented during initialization)
	if got := cache.Get(); got != 2 {
		t.Errorf("Initial value = %v, want %v", got, 2)
	}

	// Call Update and verify that the value changes
	cache.Update()
	if got := cache.Get(); got != 3 {
		t.Errorf("After Update() = %v, want %v", got, 3)
	}
}

func TestUpdateWithError(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a value that will change when Update is called successfully
	value := 1
	failNextUpdate := false
	updateFunc := func() (int, error) {
		if failNextUpdate {
			return 0, errors.New("update failed")
		}
		value++
		return value, nil
	}

	cache := New(ctx, time.Hour, updateFunc)

	// Initial value should be 2 (1 incremented during initialization)
	if got := cache.Get(); got != 2 {
		t.Errorf("Initial value = %v, want %v", got, 2)
	}

	// Make the next update fail
	failNextUpdate = true

	// Call Update and verify that the value doesn't change
	cache.Update()
	if got := cache.Get(); got != 2 {
		t.Errorf("After failed Update() = %v, want %v", got, 2)
	}

	// Make the next update succeed
	failNextUpdate = false

	// Call Update and verify that the value changes
	cache.Update()
	if got := cache.Get(); got != 3 {
		t.Errorf("After successful Update() = %v, want %v", got, 3)
	}
}

func TestUpdateLoop(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to signal when updates occur
	updateCh := make(chan int, 10)
	updateCount := 0
	updateFunc := func() (int, error) {
		updateCount++
		updateCh <- updateCount
		return updateCount, nil
	}

	// Create a cache with a short update period
	cache := New(ctx, 50*time.Millisecond, updateFunc)

	// Wait for at least 3 updates to occur
	for i := 0; i < 3; i++ {
		select {
		case <-updateCh:
			// Update received
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Timed out waiting for update %d", i+1)
		}
	}

	// Verify that updateFunc was called multiple times
	if got := cache.Get(); got < 3 {
		t.Errorf("After waiting, value = %v, want at least 3", got)
	}

	// Cancel the context to stop the update loop
	cancel()

	// Record the current value
	currentValue := cache.Get()

	// Wait to ensure the update loop has stopped
	time.Sleep(100 * time.Millisecond)

	// Verify that the value hasn't changed
	if got := cache.Get(); got != currentValue {
		t.Errorf("After canceling context, value = %v, want %v", got, currentValue)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a thread-safe counter for the update function
	var counter int64
	updateFunc := func() (int64, error) {
		return atomic.AddInt64(&counter, 1), nil
	}

	// Create a cache
	cache := New(ctx, 10*time.Millisecond, updateFunc)

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Launch multiple goroutines to access the cache concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				// Get the current value
				_ = cache.Get()

				// Occasionally call Update
				if j%10 == 0 {
					cache.Update()
				}

				// Sleep briefly to allow other goroutines to run
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// If we got here without panicking, the test passes
}

func TestGlobalCacheUpdate(t *testing.T) {
	updateCount := 0
	updateFunc := func() (int, error) {
		updateCount++
		return updateCount, nil
	}
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		_ = New(ctx, time.Hour, updateFunc)
	}

	GlobalCacheUpdate()

	if updateCount != 20 {
		t.Errorf("Expected 20 updates, got %d", updateCount)
	}
}

func TestCacheLifecycle(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to signal when updates occur
	updateCh := make(chan int, 10)
	value := 0
	updateFunc := func() (int, error) {
		value++
		updateCh <- value
		return value, nil
	}

	// Create a cache with a short update period
	cache := New(ctx, 20*time.Millisecond, updateFunc)

	// Verify initial value
	if got := cache.Get(); got != 1 {
		t.Errorf("Initial value = %v, want %v", got, 1)
	}

	// Drain the initial update from the channel
	<-updateCh

	// Wait for at least 2 more updates to occur
	for i := 0; i < 2; i++ {
		select {
		case val := <-updateCh:
			t.Logf("Received update: %d", val)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Timed out waiting for update %d", i+1)
		}
	}

	// Verify that the value has been updated automatically
	got := cache.Get()
	t.Logf("Current value after waiting: %d", got)
	if got < 3 {
		t.Errorf("After waiting, value = %v, want at least 3", got)
	}

	// Manually update the cache
	cache.Update()

	// Record the current value
	currentValue := cache.Get()

	// Cancel the context to stop the update loop
	cancel()

	// Wait to ensure the update loop has stopped
	time.Sleep(100 * time.Millisecond)

	// Verify that the value hasn't changed automatically
	if got := cache.Get(); got != currentValue {
		t.Errorf("After canceling context, value = %v, want %v", got, currentValue)
	}

	// Manually update the cache one more time
	cache.Update()

	// Verify that manual updates still work
	if got := cache.Get(); got != currentValue+1 {
		t.Errorf("After manual update, value = %v, want %v", got, currentValue+1)
	}
}
