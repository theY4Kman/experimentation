package trie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)


func TestTrie(t *testing.T) {
    Convey("a trie", t, func() {
        Convey("should remember inserted values", func() {
            tr := NewTrie()
            tr.Set("key", "value")
            So(tr.Get("key"), ShouldEqual, "value")
        })

        Convey("should split nodes into children when same-length key is lexicographically smaller", func() {
            tr := NewTrie()
            tr.Set("cat", "cat")
            tr.Set("car", "car")
            So(tr.Get("cat"), ShouldEqual, "cat")
            So(tr.Get("car"), ShouldEqual, "car")
        })

        Convey("should split nodes into children when same-length key is lexicographically larger", func() {
            tr := NewTrie()
            tr.Set("car", "car")
            tr.Set("cat", "cat")
            So(tr.Get("car"), ShouldEqual, "car")
            So(tr.Get("cat"), ShouldEqual, "cat")
        })

        Convey("should add a tail node when key is lexicographically larger", func() {
            tr := NewTrie()
            tr.Set("car", "car")
            tr.Set("cad", "cad")
            tr.Set("cat", "cat")
            So(tr.Get("car"), ShouldEqual, "car")
            So(tr.Get("cad"), ShouldEqual, "cad")
            So(tr.Get("cat"), ShouldEqual, "cat")
        })

        Convey("should overwrite the value of a key inserted twice", func() {
            tr := NewTrie()
            tr.Set("moo", "old")
            tr.Set("moo", "moo")
            So(tr.Get("moo"), ShouldEqual, "moo")
        })

        Convey("should traverse child nodes with consecutively longer keys", func() {
            tr := NewTrie()
            tr.Set("car", "car")
            tr.Set("cart", "cart")
            tr.Set("carts", "carts")
            So(tr.Get("car"), ShouldEqual, "car")
            So(tr.Get("cart"), ShouldEqual, "cart")
            So(tr.Get("carts"), ShouldEqual, "carts")
        })

        Convey("should add child nodes when keys share common prefix", func() {
            tr := NewTrie()
            tr.Set("kill", "kill")
            tr.Set("kills", "kills")
            tr.Set("killed", "killed")
            So(tr.Get("kill"), ShouldEqual, "kill")
            So(tr.Get("kills"), ShouldEqual, "kills")
            So(tr.Get("killed"), ShouldEqual, "killed")
        })
    })
}
