## TODO

- [ ] Notifier `lastFired` is updated by `Send` on the published snapshot while `Clone` may be reading it. Move backoff state out of `Config` (e.g. a notifier registry keyed by name), or guard it
- [ ] Subscribers run while holding `wmu` (see TODO in `Subscribe`); the scheduler's subscriber blocks until its loop accepts the new schedule
- [ ] Keep `Config.Clone` in sync when adding pointer/slice/map fields to `Config`, `Service`, or any `*Trigger` (covered by `TestConfigCloneIsDeep`)
- [ ] Refactor GC (garbage collector) to handle renames such that GC on rename does not clear history for that service.
