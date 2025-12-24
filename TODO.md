# TODO

This file outlines potential areas for improvement in the `goUp` project.

## Frontend #1
- [ ] Add features like sorting, filtering, and searching for services.
    - [x] Searching
    - [ ] Sorting
    - [ ] Filtering
- [ ] Consider adding a dark mode to the UI.
- [ ] Make the frontend "look good"

## Extensibility #2
- [ ] Add more notification triggers, such as:
    - [ ] Email notifications
    - [ ] Slack notifications
- [ ] Make the trigger system even more modular to allow users to easily add their own custom triggers.

## Error Handling #3
- [ ] Implement a more sophisticated error handling and retry mechanism for service checks.
    - For example, instead of immediately marking a service as down after one failed check, allow for a configurable number of retries.
- [ ] Add more detailed error messages to the frontend to help with debugging.
    - [/] Removing fatal errors and panics and handling errors to frontend

## Configuration #4
- [ ] For projects with many services, the `services.yml` file could become hard to manage. Consider adding a UI or a CLI to manage the list of services.
- [ ] Add support for dynamic configuration reloading without restarting the application.

## Testing #5
- [ ] Add unit tests for the Go backend, especially for the utility functions (`utils/*.go`).
- [ ] Add integration tests to ensure the different parts of the application (scheduler, data store, server) work correctly together.
- [ ] Add frontend tests to ensure the UI displays the data correctly.

## CI/CD #6
- [ ] Implement a CI/CD pipeline to automate testing and builds.