# i18n [![GoDoc](https://godoc.org/github.com/worldiety/i18n?status.svg)](http://godoc.org/github.com/worldiety/i18n)

The API is entirely developer
centric and not build around any exchange format like json, toml or xml. This is intentional and based
on our observation that most projects are not translated anymore in a classical way. With the rising
of various AI tools and their reached level of quality, there is "no" need anymore for an average app to get
translated by professionals. This is mostly driven by the customer and the market expectations that
developers should just "start over" using those tools and weired translations are fixed only on demand
if they are found.

Thus, this API uses the type system to remove redundant declaration and notation work from the developer
and offloads it to the compiler, where possible without code generation.
We expect that translation fixes and additional languages are
added later at runtime and if not already written by the developer, it will be added by end users using
some sort of user interface without the help of any developer or professional translator.

Note also that the implementation is optimized for speed and therefore brings its own parser and
parameter notation so that evaluation can be as fast as possible which is not true for most other
implementations. We care especially for a lock free lookup of quasi constant strings, which is the
most common thing in our Apps. Also, any value string reference is always int-sized, and thus it is half the size of
a regular string (fat) pointer. This also guarantees a real O(1) index lookup for any key in a localized bundle.
Most other implementations use string keys, which must always be hashed and be found in a hashtable (collisions, 
equals etc.) - our approach is magnitudes more efficient than that. We also cache the pre-parsed variable strings
and optimize (re-)allocation where possible.

There is also a convenient string-only [Bundle.Resolve] support 
to tunnel string resource handles through arbitrary string. Even though this looks slow, it still provides 
acceptable performance.

This library was made for https://github.com/worldiety/nago but can be used freely in other projects.