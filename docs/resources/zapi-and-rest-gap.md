# Zapi and Rest gaps

## Volume Count difference
`REST` and `ZAPI` collector will return different counts of volume object depending on whether you have object store servers setup on your cluster or not.

- `REST` collector would return all the volumes but object store servers volumes.
- `ZAPI` collector would return all the volumes including object store servers volumes.