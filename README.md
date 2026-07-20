# shiftbell

Self-hosted chore tracker for a homelab.

Currently pre-v1 and under development.

## features

- Create and complete one-off chores.
- Define recurring chores, where completion automatically creates the next one.
- Keep a completed chore history.

## audience

Primarily a personal app, designed to be self-hosted in a homelab.

## stack

Go+htmx and sqlite, packaged as a single binary.

## quick start

TODO: See [development](#development) for now.

## development

Use the `justfile` (see [just](https://just.systems/man/en/installation.html)) recipes. The main command is:

```sh
just up
```

Other useful recipes include `just down`, `just fmt`, and `just test`.
