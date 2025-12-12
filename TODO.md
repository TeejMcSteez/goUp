# TODO

This file outlines potential areas for improvement in the `goUp` project.

## Data Persistence #1
- [x] Replace the in-memory data store with a database like SQLite or PostgreSQL to persist historical uptime data.
- [x] Uptime percentage calculations and long-term performance monitoring from database.

## Frontend #2/3
- [ ] Improve the user experience of the frontend
- [ ] Add features like sorting, filtering, and searching for services.
- [x] Display historical uptime data for each service in a chart or graph.
- [ ] Consider adding a dark mode to the UI.

## Extensibility #2/3
- [ ] Add more notification triggers, such as:
    - [ ] Email notifications
    - [ ] Slack notifications
    - [x] Webhook notifications
- [ ] Make the trigger system even more modular to allow users to easily add their own custom triggers.

## Error Handling #4
- [ ] Implement a more sophisticated error handling and retry mechanism for service checks.
    - For example, instead of immediately marking a service as down after one failed check, allow for a configurable number of retries.
- [ ] Add more detailed error messages to the frontend to help with debugging.

## Configuration #5
- [ ] For projects with many services, the `services.yml` file could become hard to manage. Consider adding a UI or a CLI to manage the list of services.
- [ ] Add support for dynamic configuration reloading without restarting the application.

## Testing #6
- [ ] Add unit tests for the Go backend, especially for the utility functions (`utils/*.go`).
- [ ] Add integration tests to ensure the different parts of the application (scheduler, data store, server) work correctly together.
- [ ] Add frontend tests to ensure the UI displays the data correctly.

## CI/CD #7
- [ ] Implement a CI/CD pipeline to automate testing and builds.