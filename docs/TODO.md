## TODO

- [/] Notifier `lastFired` is updated by `Send` on the published snapshot while `Clone` may be reading it. Move backoff state out of `Config` (e.g. a notifier registry keyed by name), or guard it
  - Issue where race can still occur if 2 Send() fire at the same time and pass check before either is marked. This is unlikely using my loop but if work keeps getting split up could cause issues.
- [ ] Subscribers run while holding `wmu` (see TODO in `Subscribe`); the scheduler's subscriber blocks until its loop accepts the new schedule
- [ ] Keep `Config.Clone` in sync when adding pointer/slice/map fields to `Config`, `Service`, or any `*Trigger` (covered by `TestConfigCloneIsDeep`)
- [ ] Refactor GC (garbage collector) to handle renames such that GC on rename does not clear history for that service.
- [ ] Allow configurable query for `host online` checks for people that block specific servers.
- [ ] Implement config service update over Websocket
- [ ] Handle chunking large Discord messages (2000 word limit)
- [ ] factor time parsing out of `utils/datastore.go` to a general helper in `utils/utils.go` to be used in `server/handlers_websocket.go` and `utils/datastore.go`
