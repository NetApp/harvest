## v1.4.0

- Add `SetExpires` method to match `Section` interface and match naming scheme of the other expire-related endpoints.  
  → Hence, **`SetExpire` is now deprecated and will be removed in the next version.** Please use `SetExpires` instead.

- Make use of `sync.Pool` to re-use `element` instances instead of creating new ones on each new element creation and passing them to the GC after deletion.

- Add `StartCleanerInternal` and `StartCleanerExternal` endpoints to be able to re-start the internal cleanup loop with new specifications.

## v1.3.1

- Fix `concurrent map read and map write` panic when accessing the map concurrently while getting an existing key, which is expired at the time. [#4]