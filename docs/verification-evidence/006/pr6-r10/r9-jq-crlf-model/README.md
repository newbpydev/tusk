# R10 retained CRLF model prerequisites

The retained R10 wrapper ran with `jq-1.8.2` on the declared Linux host. That
jq executable accepts `--binary`, which the wrapper uses to obtain LF before
modeling native Windows default CRLF output. Direct wrapper calls require an
explicit `MODEL_REAL_JQ` executable selector. Select and probe that same binary:

```bash
export MODEL_REAL_JQ="$(command -v jq)"
"$MODEL_REAL_JQ" --binary --null-input empty
```

A nonzero status means the selected jq cannot run this particular historical
control. The sanctioned `r10-jq-crlf-model.sh` replay recipe selects the PATH jq
before prepending the wrapper directory and passes that selector to its children.
Use the recipe with `TUSK_CHECKOUT` set to the owned checkout root; invoking the
wrapper directly also requires the exported selector shown above.
The `--binary` capability is a prerequisite of this historical control,
not of Tusk's Unix maintainer scripts, which omit the Windows-only flag.

The separate retained legacy Unix control rejects that option and checks the
supported Unix script behavior. Original wrapper bytes and measured logs are
preserved; this note records the previously unstated model-tool identity.
