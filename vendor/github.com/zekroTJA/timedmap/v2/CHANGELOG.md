## v2.0.0

- Add type parameters for the key and value type of a TimedMap and corresponding constructor functions.
- Remove the section system for better simplicity and usability.
- Remove deprecated `SetExpire` method.
- Update documentation.
- Update minimum required Go version to v1.19.0.

## v1.5.2

- Multiple race conditions have been fixed (by @ShivamKumar2002 in https://github.com/zekroTJA/timedmap/pull/8)

## v1.5.1

- Add [`FromMap`](https://pkg.go.dev/github.com/zekroTJA/timedmap#FromMap) constructor which can be used to create a `TimedMap` from an existing map with the given expiration values for each key-value pair.

## v1.4.0

- Add `SetExpires` method to match `Section` interface and match naming scheme of the other expire-related endpoints.  
  → Hence, **`SetExpire` is now deprecated and will be removed in the next version.** Please use `SetExpires` instead.

- Make use of `sync.Pool` to re-use `element` instances instead of creating new ones on each new element creation and passing them to the GC after deletion.

- Add `StartCleanerInternal` and `StartCleanerExternal` endpoints to be able to re-start the internal cleanup loop with new specifications.

## v1.3.1

- Fix `concurrent map read and map write` panic when accessing the map concurrently while getting an existing key, which is expired at the time. [#4]