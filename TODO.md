# TODO

This file outlines potential areas for improvement in the `goUp` project.

## Frontend
- [ ] Add features like sorting, filtering, and searching for services.
    - [ ] Sorting
    - [ ] Filtering
- [ ] Consider adding a dark mode to the UI.
- [ ] Make the frontend "look good"

## Database

- [ ] Add error field to tell the frontend explicitly whether or not their was an error rather than checking for 200 response on the frontend

## Testing
- [ ] Add unit tests for the Go backend, especially for the utility functions (`utils/*.go`).
- [ ] Add integration tests to ensure the different parts of the application (scheduler, data store, server) work correctly together.
- [ ] Add frontend tests to ensure the UI displays the data correctly.

## CI/CD
- [ ] Implement a CI/CD pipeline to automate testing and builds.