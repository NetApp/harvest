# ZAPI and REST Gaps

## Volume Count difference
The `REST` and `ZAPI` collectors return a different number of `volume_labels` depending on whether you have set up object store servers on your cluster.

- The `REST` collector does not include `volume_labels` for volumes associated with object store servers.
- The `ZAPI` collector includes `volume_labels` for volumes associated with object store servers. If you have not set up any object store servers on your cluster, both collectors will return the same number of `volume_labels`.
