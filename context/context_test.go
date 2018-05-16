// Includes code from Gorilla context test:
// https://github.com/gorilla/context/blob/master/context_test.go
// © 2012 The Gorilla Authors
package context

import (
	"testing"

	"github.com/lytics/gentleman/utils"
)

type keyType int

const (
	key1 keyType = iota
	key2
)

func TestContext(t *testing.T) {
	ctx := New()
	store := ctx.getStore()

	// Get()
	utils.Equal(t, ctx.Get(key1), nil)

	// Set()
	ctx.Set(key1, "1")
	utils.Equal(t, ctx.Get(key1), "1")
	utils.Equal(t, len(store), 1)
	utils.Equal(t, store[key1], "1")

	ctx.Set(key2, "2")
	utils.Equal(t, ctx.Get(key2), "2")
	utils.Equal(t, len(store), 2)

	// GetOk()
	value, ok := ctx.GetOk(key1)
	utils.Equal(t, value, "1")
	utils.Equal(t, ok, true)

	value, ok = ctx.GetOk("not exists")
	utils.Equal(t, value, nil)
	utils.Equal(t, ok, false)

	ctx.Set("nil value", nil)
	value, ok = ctx.GetOk("nil value")
	utils.Equal(t, value, nil)
	utils.Equal(t, ok, true)

	// GetString()
	ctx.Set("int value", 13)
	ctx.Set("string value", "hello")
	str := ctx.GetString("int value")
	utils.Equal(t, str, "")
	str = ctx.GetString("string value")
	utils.Equal(t, str, "hello")

	// GetAll()
	values := ctx.GetAll()
	utils.Equal(t, len(values), 5)

	// Delete()
	ctx.Delete(key1)
	utils.Equal(t, ctx.Get(key1), nil)
	utils.Equal(t, len(store), 4)

	ctx.Delete(key2)
	utils.Equal(t, ctx.Get(key2), nil)
	utils.Equal(t, len(store), 3)

	// Clear()
	ctx.Set(key1, true)
	values = ctx.GetAll()
	ctx.Clear()
	utils.Equal(t, len(store), 0)
	val, _ := values["int value"].(int)
	utils.Equal(t, val, 13) // Clear shouldn't delete values grabbed before
}

func TestContextInheritance(t *testing.T) {
	parent := New()
	ctx := New()
	ctx.UseParent(parent)

	parent.Set("foo", "bar")
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.Get("foo"), "bar")
	utils.Equal(t, ctx.Get("bar"), "foo")

	ctx.Set("foo", "foo")
	utils.Equal(t, ctx.Get("foo"), "foo")
}

func TestContextGetAll(t *testing.T) {
	parent := New()
	ctx := New()
	ctx.UseParent(parent)

	parent.Set("foo", "bar")
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.Get("foo"), "bar")
	utils.Equal(t, ctx.Get("bar"), "foo")

	store := ctx.GetAll()
	utils.Equal(t, len(store), 2)
}

func TestContextRoot(t *testing.T) {
	root := New()
	parent := New()
	parent.UseParent(root)
	ctx := New()
	ctx.UseParent(parent)
	if ctx.Root() != root {
		t.Error("Invalid root context")
	}
}

func TestContextGetters(t *testing.T) {
	parent := New()
	ctx := New()
	ctx.UseParent(parent)

	parent.Set("foo", "bar")
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.GetString("foo"), "bar")
	utils.Equal(t, ctx.GetString("bar"), "foo")
	ctx.Clear()

	parent.Set("foo", 1)
	ctx.Set("bar", 2)
	foo, ok := ctx.GetInt("foo")
	utils.Equal(t, ok, true)
	utils.Equal(t, foo, 1)
	bar, ok := ctx.GetInt("bar")
	utils.Equal(t, ok, true)
	utils.Equal(t, bar, 2)

	store := ctx.GetAll()
	utils.Equal(t, len(store), 2)
}

func TestContextClone(t *testing.T) {
	ctx := New()
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.Get("bar"), "foo")

	newCtx := ctx.Clone()
	utils.Equal(t, newCtx.Get("bar"), "foo")

	// Overwrite value
	newCtx.Set("bar", "bar")
	utils.Equal(t, ctx.Get("bar"), "foo")
	utils.Equal(t, newCtx.Get("bar"), "bar")
}

func TestContextCopy(t *testing.T) {
	ctx := New()
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.Get("bar"), "foo")

	newCtx := New()
	ctx.CopyTo(newCtx)
	utils.Equal(t, newCtx.Get("bar"), "foo")

	// Ensure inmutability
	newCtx.Set("bar", "bar")
	utils.Equal(t, ctx.Get("bar"), "foo")
	utils.Equal(t, newCtx.Get("bar"), "bar")
}

func TestContextSetRequest(t *testing.T) {
	ctx := New()
	ctx.Set("bar", "foo")
	utils.Equal(t, ctx.Get("bar"), "foo")

	newCtx := New()
	ctx.CopyTo(newCtx)
	utils.Equal(t, newCtx.Get("bar"), "foo")

	// Ensure inmutability
	newCtx.Set("bar", "bar")
	utils.Equal(t, ctx.Get("bar"), "foo")
	utils.Equal(t, newCtx.Get("bar"), "bar")
}
