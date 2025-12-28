# TODO

This file outlines potential areas for improvement in the `goUp` project.

## Frontend
- [ ] Add features like sorting, filtering, and searching for services.
    - [ ] Sorting
    - [ ] Filtering
- [ ] Consider adding a dark mode to the UI.
- [ ] Make the frontend "look good"

## Extensibility
- [ ] For ease of development and to ensure I can maintain this project, I will be doing MQTT triggers and/or Webhook triggers as most emails, slack messages, etc. can be fired from webhooks and MQTT is a light protocol for simple messaging.
    - [ ] Increase modularity of webhook Fire
        - [ ] Custom Webhook Messages
            - [ ] Create a message crafter to make a custom message to send to the trigger. Mainly for things like Email and Slack that take in custom messages.

## Error Handling
- [ ] Implement a more sophisticated error handling and retry mechanism for service checks.
    - For example, instead of immediately marking a service as down after one failed check, allow for a configurable number of retries.
- [ ] Add more detailed error messages to the frontend to help with debugging.

## Testing
- [ ] Add unit tests for the Go backend, especially for the utility functions (`utils/*.go`).
- [ ] Add integration tests to ensure the different parts of the application (scheduler, data store, server) work correctly together.
- [ ] Add frontend tests to ensure the UI displays the data correctly.

## CI/CD
- [ ] Implement a CI/CD pipeline to automate testing and builds.