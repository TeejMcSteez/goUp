## TODO

### ConfigStore migration

Goal: remove the global `utils.Current_Config` (and `configMu`) and make `utils.ConfigStore` the only source of config. Reads go through `store.Get()`, writes through `store.Update(func(c *Config) error)`, and side effects (Setup, reschedule, WS broadcast) run as subscribers.

**Done**
- [x] `Current_Config`, `configMu`, `writeConfig`, `LoadConfig`, `ConfigWriteNotify`/`programmaticWrite` removed
- [x] `utils/conf.go` mutators edit a `*Config` in place and no longer write to disk; callers wrap them in `store.Update`
- [x] Config-reading helpers take the snapshot as a parameter: `InitDB`, `GetDatabaseSize`, `CleanupDbFiles`, `GetMaxSize`, `GetRecentData`, `GetLiveData`, `Check`, `GetServiceData`, `Get*Uptime` / `GetUptimeAverage`
- [x] `server`: `Server.store`; config/db/data/websocket handlers read `s.store.Get()` once per request and write via `s.edit` → `store.Update`
- [x] `ConfigStore.Update` / `LoadFromFile` / `Reload` run `SetupTrigger` on the private copy before publishing (replaces the old `writeConfig` refresh, #63)
- [x] `workers/scheduler.go`: one `store.Get()` snapshot per fetch cycle; `Scheduler.Update` persists via the store and the scheduler subscribes itself to schedule changes
- [x] `workers/hot.go`: `store.Reload()` skips its own writes by hash (`lastWritten`, read under `wmu`) and runs subscribers. A bad edit no longer stops the reloader
- [x] `workers/memory.go`: reads max size and db path from the store every tick
- [x] Tests: no test sets a global config any more
- [x] `utils/scanner.go`: stateless `Scanner` derives endpoints from each config snapshot (`Endpoints(cfg)`). `services.go`, `svcEndpoints`, `Setup`, the startup Setup call and the services subscriber are gone. Tests put endpoints in the config's `Services` instead of a global

**Remaining**
- [ ] Notifier `lastFired` is updated by `Send` on the published snapshot while `Clone` may be reading it. Move backoff state out of `Config` (e.g. a notifier registry keyed by name), or guard it
- [ ] `UpdateConfigService` / `DeleteConfigService` touch the DB inside the `Update` closure (under the store lock, and before the save can fail). Move the DB work into the handler after a successful `Update`
- [ ] `DbGarbageCollect` is still called manually (handlers, hot reloader, startup) rather than from a subscriber. A subscriber would run before the handler's DB rename, so ordering needs thought first
- [ ] Subscribers run while holding `wmu` (see TODO in `Subscribe`); the scheduler's subscriber blocks until its loop accepts the new schedule
- [ ] Keep `Config.Clone` in sync when adding pointer/slice/map fields to `Config`, `Service`, or any `*Trigger` (covered by `TestConfigCloneIsDeep`)
