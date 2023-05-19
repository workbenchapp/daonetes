## Generated Go bindings

For instances where upstream does not directly support built-in Solana
programs, or where we have our own program code via Anchor, we must
generate the client bindings from an Anchor IDL.

to generate the bindings for worknet:

```
$ ./bindings.sh
```

Bindings for built-in programs don't change much, so
to generate bindings for built-in programs you must set
the `ANCHOR_CODEGEN_ALL` environment variable:

```
$ ANCHOR_CODEGEN_ALL ./bindings.sh
```
