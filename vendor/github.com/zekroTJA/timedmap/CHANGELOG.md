## v1.3.1

- Fix `concurrent map read and map write` panic when accessing the map concurrently while getting an existing key, which is expired at the time. [#4]