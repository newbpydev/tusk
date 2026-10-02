# R10 retained CRLF model prerequisites

The retained R10 wrapper ran with `jq-1.8.2` on the declared Linux host. That
jq executable accepts `--binary`, which the wrapper uses to obtain LF before
modeling native Windows default CRLF output. Run `jq --binary --null-input empty`
before replaying this particular retained model; a nonzero status means the host
jq cannot run it. This capability is a prerequisite of this historical control,
not of Tusk's Unix maintainer scripts, which omit the Windows-only flag.

The separate retained legacy Unix control rejects that option and checks the
supported Unix script behavior. Original wrapper bytes and measured logs are
preserved; this note records the previously unstated model-tool identity.
