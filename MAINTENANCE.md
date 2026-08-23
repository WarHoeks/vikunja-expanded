# Maintenance

## Syncing with upstream

1. `git fetch upstream`
2. `git checkout upstream-main && git merge --ff-only upstream/main && git push origin upstream-main`
3. `git checkout main && git merge upstream-main`
4. Resolve conflicts (see "Touch points" below for where they're likely to show up), then build and test locally before pushing.
5. `git push origin main` — CI builds and publishes the new image automatically.

Do this whenever you want a specific new upstream feature or fix, not on a fixed schedule. Small, frequent syncs are much easier to resolve than letting months of upstream changes pile up.

Before deploying a synced release to the real stack, snapshot the Postgres volume (`vikunja-dbdata`) in case a migration ordering issue surfaces.

## Touch points

Files this fork edits directly (as opposed to new files it only adds). These are the ones to watch for conflicts on every sync:

| File | What was changed | Why |
|---|---|---|
| `pkg/models/project.go` | Added `DefaultTaskFields` field | Backs the "default task fields per project" QoL feature |
| `frontend/src/components/project/ProjectWrapper.vue` (path to confirm) | Wraps the existing view output + new `<ProjectInfoSidebar>` in a flex row | Adds the resizable info/links sidebar |
| Task detail "Management" panel component (path to confirm) | Added an "Add web link" button next to "Add attachments" | Task web links feature |
| Task detail field-visibility logic (path to confirm) | `v-if`s for optional fields OR'd with the project's default-fields list | Default task fields QoL feature |
| Labels multiselect component (path to confirm) | Shows full label list on focus instead of only after typing | Labels QoL feature |

Everything else this fork adds is new files (new Go models/migrations/v2 API resource files, new Vue components), which don't produce merge conflicts on their own.

## Notes

- New backend endpoints target the **v2 API** (`pkg/routes/api/v2/`, Huma-based, self-registering via `init()` + `AddRouteRegistrar()`), not v1 — v1 is deprecated in Vikunja 3.0 and removed in 4.0. At each sync, sanity-check `pkg/routes/api/v2/registry.go` and one existing resource file (e.g. `task_attachments.go`) still look the same shape.
- This repo inherited a bunch of upstream-only GitHub Actions workflows (crowdin, auto-label, nixpkgs-update, release.yml, stale bot, etc.) that assume upstream's secrets/infra. They're harmless no-ops here (they'll just fail auth or not trigger), but if the noise in the Actions tab is annoying, disable them individually from repo Settings → Actions rather than deleting the files, so upstream syncs don't keep re-adding them.
