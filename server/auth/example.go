package auth

// Example usage - Integration guide for server.go
//
// This file shows how to integrate the auth system into your existing server.
// DO NOT run this file directly - it's just for reference.

/*

1. First, add the bcrypt dependency to your go.mod:
   Run: go get golang.org/x/crypto/bcrypt

2. In your main function or server setup, initialize the auth database:

   authDB, err := auth.SetupStore()
   if err != nil {
       log.Fatal(err)
   }
   defer authDB.Close()

   // Create auth handler and middleware
   authHandler := auth.NewAuthHandler(authDB, false) // Set to true in production for HTTPS
   authMiddleware := auth.NewAuthMiddleware(authDB)

3. Add auth routes in your server's Start() method:

   // Public auth endpoints
   http.HandleFunc("/auth/register", authHandler.Register)
   http.HandleFunc("/auth/login", authHandler.Login)
   http.HandleFunc("/auth/logout", authHandler.Logout)
   http.HandleFunc("/auth/me", authHandler.Me)

4. Protect existing routes with middleware:

   // Example: Protect the schedule API
   http.HandleFunc("/api/schedule", authMiddleware.RequireAuth(s.ScheduleApi))

   // Example: Protect the clear database endpoint
   http.HandleFunc("/api/db/clear", authMiddleware.RequireAuth(s.ClearDatabase))

5. Optional: Use OptionalAuth for routes that work with or without auth:

   http.HandleFunc("/api", authMiddleware.OptionalAuth(s.Api))

6. Access user info in protected handlers:

   func (s *Server) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
       user, ok := auth.GetUserFromContext(r)
       if !ok {
           http.Error(w, "Unauthorized", http.StatusUnauthorized)
           return
       }

       log.Printf("Request from user: %s (ID: %d)", user.Username, user.ID)
       // ... rest of handler
   }

7. Optional: Setup periodic session cleanup (run in a goroutine):

   go func() {
       ticker := time.NewTicker(1 * time.Hour)
       defer ticker.Stop()
       for range ticker.C {
           auth.CleanupExpiredSessions(authDB)
       }
   }()

8. Frontend integration examples:

   // Register
   fetch('/auth/register', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       body: JSON.stringify({ username: 'user', password: 'password123' }),
       credentials: 'include' // Important: includes cookies
   })

   // Login
   fetch('/auth/login', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       body: JSON.stringify({ username: 'user', password: 'password123' }),
       credentials: 'include'
   })

   // Check authentication
   fetch('/auth/me', {
       credentials: 'include'
   })

   // Logout
   fetch('/auth/logout', {
       method: 'POST',
       credentials: 'include'
   })

   // Make authenticated requests
   fetch('/api/schedule', {
       credentials: 'include'
   })

9. Security checklist for production:

   ✓ Set authHandler secure flag to true (enables HTTPS-only cookies)
   ✓ Use HTTPS in production
   ✓ Set strong password requirements in password.go
   ✓ Implement rate limiting on login/register endpoints
   ✓ Add CORS configuration if frontend is on different domain
   ✓ Monitor failed login attempts
   ✓ Regular session cleanup
   ✓ Consider adding 2FA for sensitive operations
   ✓ Add logging for security events
   ✓ Review and test all authentication flows

10. Testing:

   // Create a test user
   curl -X POST http://localhost:8101/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass123"}' \
     -c cookies.txt

   // Login (saves cookie)
   curl -X POST http://localhost:8101/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass123"}' \
     -c cookies.txt -b cookies.txt

   // Check auth status
   curl http://localhost:8101/auth/me \
     -b cookies.txt

   // Access protected endpoint
   curl http://localhost:8101/api/schedule \
     -b cookies.txt

   // Logout
   curl -X POST http://localhost:8101/auth/logout \
     -b cookies.txt

*/
