### ibdf-go

Reads and writes network packages to a chunk based [piff](https://github.com/piot/piff-go) file format.

For each network packet it stores the:

- Direction (incoming or outgoing)
- Timestamp
- the actual packet octets.

The start of the file also contains schema octets that are implementation specific.
