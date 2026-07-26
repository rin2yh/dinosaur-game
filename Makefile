# Every command is defined once in mise.toml. These targets only
# delegate to it so `make <target>` and `mise run <task>` cannot drift.

TASKS := build build-web build-koebiten flash-koebiten run serve fmt lint test check clean

.PHONY: all $(TASKS)

all: check build-web

$(TASKS):
	mise run $@
