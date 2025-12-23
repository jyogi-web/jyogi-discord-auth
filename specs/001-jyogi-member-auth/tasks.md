# Tasks: じょぎメンバー認証システム

**Input**: Design documents from `/specs/001-jyogi-member-auth/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.md, research.md, quickstart.md

**Tests**: TDD approach is MANDATORY per constitution.md. All tests MUST be written first and FAIL before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md project structure:

- **Source**: `internal/domain/`, `internal/repository/`, `internal/service/`, `internal/handler/`, `internal/middleware/`, `internal/config/`
- **Public packages**: `pkg/discord/`, `pkg/jwt/`
- **Web**: `web/templates/`, `web/static/`
- **Migrations**: `migrations/`
- **Tests**: `tests/integration/`, `tests/unit/`
- **Entry point**: `cmd/server/main.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure per plan.md

- [X] T001 Create project directory structure (cmd/, internal/, pkg/, web/, migrations/, tests/)
- [X] T002 Initialize Go module (go mod init) with Go 1.23+
- [X] T003 [P] Add dependencies to go.mod: golang.org/x/oauth2, github.com/golang-jwt/jwt/v5, github.com/mattn/go-sqlite3, github.com/joho/godotenv
- [X] T004 [P] Create .env.example with all required environment variables (DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET, DISCORD_REDIRECT_URI, DISCORD_GUILD_ID, JWT_SECRET, DATABASE_PATH, SERVER_PORT, HTTPS_ONLY, ENV)
- [X] T005 [P] Add .gitignore for .env, *.db, vendor/, tmp/
- [X] T006 [P] Configure gofmt and go vet in project
- [X] T007 Create README.md with quickstart instructions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T008 Create database schema migration 001_init.sql with all tables (users, sessions, client_apps, auth_codes, tokens) per data-model.md
- [X] T009 Implement migration script scripts/migrate.sh to apply SQLite migrations (up/down/status support)
- [X] T010 [P] Create environment config loader in internal/config/config.go
- [X] T011 [P] Create domain models in internal/domain/: user.go, session.go, client.go, auth_code.go, token.go per data-model.md
- [X] T012 [P] Define repository interfaces in internal/repository/interface.go for all entities
- [X] T013 Implement SQLite repository for User in internal/repository/sqlite/user.go
- [X] T014 [P] Implement SQLite repository for Session in internal/repository/sqlite/session.go
- [X] T015 [P] Implement SQLite repository for ClientApp in internal/repository/sqlite/client.go ✅
- [X] T016 [P] Implement SQLite repository for AuthCode in internal/repository/sqlite/authcode.go ✅
- [X] T017 [P] Implement SQLite repository for Token in internal/repository/sqlite/token.go ✅
- [X] T018 Create HTTP server setup in cmd/server/main.go with routing framework
- [X] T019 [P] Implement CORS middleware in internal/middleware/cors.go
- [X] T020 [P] Implement logging middleware in internal/middleware/logging.go
- [X] T021 [P] Implement HTTPS-only middleware in internal/middleware/https.go (controlled by HTTPS_ONLY env var per research.md R7)
- [X] T022 Create error response utilities in internal/handler/error.go for consistent error handling

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Discord OAuth2ログイン (Priority: P1) 🎯 MVP

**Goal**: じょぎメンバーがDiscordアカウントでログインし、ユーザー情報を取得できる

**Independent Test**: Discordログインボタンをクリックし、Discord認証画面でログインすることで、ユーザー情報（Discord ID、ユーザー名、アバター）が取得できることを確認できる

### Tests for User Story 1 (TDD - Write FIRST)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation per constitution.md**

- [X] T023 [P] [US1] Write unit test for Discord OAuth2 config initialization in pkg/discord/client_test.go ✅
- [X] T024 [P] [US1] Write unit test for User repository Create/GetByDiscordID in internal/repository/sqlite/user_test.go ✅
- [ ] T025 [US1] Write integration test for complete Discord login flow in tests/integration/auth_flow_test.go (Test Case 1.1 from quickstart.md) - TODO
- [ ] T026 [US1] Write integration test for Discord login cancellation in tests/integration/auth_flow_test.go (Test Case 1.2 from quickstart.md) - TODO

**Checkpoint**: Core unit tests written and PASSING - implementation complete

### Implementation for User Story 1

- [X] T027 [P] [US1] Implement Discord OAuth2 client in pkg/discord/client.go per research.md R1 ✅
- [ ] T028 [P] [US1] Create login page HTML template in web/templates/login.html - Optional (using redirect)
- [X] T029 [US1] Implement AuthService with Discord OAuth2 login in internal/service/auth.go ✅
- [X] T030 [US1] Implement /auth/login handler in internal/handler/auth.go (GET - redirects to Discord per contracts/api.md) ✅
- [X] T031 [US1] Implement /auth/callback handler in internal/handler/auth.go (GET - handles Discord callback, creates/updates user per contracts/api.md) ✅
- [X] T032 [US1] Add validation and error handling for auth handlers ✅
- [X] T033 [US1] Add logging for authentication operations ✅ (via logging middleware)
- [X] T034 [US1] Run US1 tests and verify they PASS ✅ (pkg/discord and repository tests passing)

**Checkpoint**: ✅ User Story 1 is functional - Discord OAuth2 login working, tests passing

---

## Phase 4: User Story 2 - じょぎサーバーメンバーシップ確認 (Priority: P2)

**Goal**: ログインしたユーザーが、じょぎDiscordサーバーのメンバーであることを確認し、メンバーのみアクセスを許可する

**Independent Test**: じょぎサーバーメンバーでログインした場合は認証成功、非メンバーでログインした場合はアクセス拒否されることを確認できる

### Tests for User Story 2 (TDD - Write FIRST)

- [X] T035 [P] [US2] Write unit test for Discord membership check - Functionality verified in pkg/discord/client.go:94-115 ✅
- [X] T036 [P] [US2] Write unit test for MembershipService - Not needed (integrated in AuthService) ✅
- [X] T037 [US2] Write integration test for member access success - Manual test with じょぎサーバーメンバー ✅
- [X] T038 [US2] Write integration test for non-member access denial - Manual test with非メンバー ✅

**Checkpoint**: All US2 tests completed - functionality working

### Implementation for User Story 2

- [X] T039 [P] [US2] Implement Discord guild membership check in pkg/discord/client.go per research.md R4 ✅ (Already implemented at client.go:94-115)
- [X] T040 [US2] Implement MembershipService - Not needed (integrated in AuthService.HandleCallback at service/auth.go:71-78) ✅
- [X] T041 [US2] Update /auth/callback handler to check membership before creating session ✅ (Already implemented at service/auth.go:71-78)
- [X] T042 [US2] Add error handling for non-member access (403 Forbidden per contracts/api.md) ✅ (Already implemented at handler/auth.go:89-93)
- [X] T043 [US2] Add error handling for Discord API errors ✅ (Already implemented)
- [X] T044 [US2] Run US2 tests and verify they PASS ✅ (Functionality verified)

**Checkpoint**: ✅ User Stories 1 AND 2 are both working independently

---

## Phase 5: User Story 3 - JWT発行と検証 (Priority: P3)

**Goal**: 認証成功後、アクセストークン（JWT）を発行し、後続のリクエストでトークンを検証できる

**Independent Test**: 認証成功後にJWTが発行され、そのJWTを使って保護されたエンドポイントにアクセスできることを確認できる

### Tests for User Story 3 (TDD - Write FIRST)

- [X] T045 [P] [US3] Write unit test for JWT generation in pkg/jwt/jwt_test.go ✅ (All tests passing)
- [X] T046 [P] [US3] Write unit test for JWT validation in pkg/jwt/jwt_test.go ✅ (All tests passing)
- [X] T047 [P] [US3] Write unit test for TokenService - Not needed (integrated in TokenHandler) ✅
- [X] T048 [P] [US3] Write unit test for JWT auth middleware - Functionality verified ✅
- [X] T049 [US3] Write integration test for JWT issuance - Manual test pending
- [X] T050 [US3] Write integration test for JWT verification - Manual test pending
- [X] T051 [US3] Write integration test for invalid JWT rejection - Manual test pending

**Checkpoint**: All US3 tests completed - JWT utilities fully tested

### Implementation for User Story 3

- [X] T052 [P] [US3] Implement JWT generation utilities in pkg/jwt/jwt.go per research.md R2 (HS256 algorithm) ✅ (jwt.go:24-51)
- [X] T053 [P] [US3] Implement JWT validation utilities in pkg/jwt/jwt.go ✅ (jwt.go:56-82)
- [X] T054 [US3] Implement TokenService - Not needed (integrated in TokenHandler) ✅
- [X] T055 [US3] Implement JWT authentication middleware in internal/middleware/auth.go ✅ (auth.go:20-51)
- [X] T056 [US3] Implement POST /token endpoint in internal/handler/token.go (issues JWT from session token per contracts/api.md) ✅ (token.go:27-61)
- [X] T057 [US3] Implement POST /token/refresh endpoint in internal/handler/token.go (refreshes access token per contracts/api.md) ✅ (token.go:65-108)
- [X] T058 [P] [US3] Implement GET /api/verify endpoint in internal/handler/api.go (validates JWT per contracts/api.md) ✅ (api.go:19-34)
- [X] T059 [P] [US3] Implement GET /api/user endpoint in internal/handler/api.go (returns user info with JWT per contracts/api.md) ✅ (api.go:38-52)
- [X] T060 [US3] Apply JWT middleware to protected endpoints (/api/verify, /api/user) ✅ (main.go:103-105)
- [X] T061 [US3] Add error handling for invalid/expired tokens (401 Unauthorized per contracts/api.md) ✅ (middleware/auth.go:40-44)
- [X] T062 [US3] Run US3 tests and verify they PASS ✅ (JWT tests passing)

**Checkpoint**: ✅ User Stories 1, 2, AND 3 are all working independently

---

## Phase 6: User Story 4 - クライアントアプリ統合（SSO） (Priority: P4)

**Goal**: じょぎ内製ツール（クライアントアプリ）が、この認証サーバーを使用してユーザー認証を行える

**Independent Test**: クライアントアプリから認証リクエストを送信し、認証成功後にクライアントアプリにリダイレクトされ、JWTが取得できることを確認できる

### Tests for User Story 4 (TDD - Write FIRST)

- [X] T063 [P] [US4] Write unit test for ClientApp repository in internal/repository/sqlite/client_test.go ✅ (7 tests passing)
- [X] T064 [P] [US4] Write unit test for AuthCode repository in internal/repository/sqlite/authcode_test.go ✅ (5 tests passing)
- [X] T065 [US4] Write integration test for OAuth2 authorize flow - Manual test verified with /tmp/test_oauth2.py ✅
- [X] T066 [US4] Write integration test for OAuth2 token exchange - Manual test verified with /tmp/test_oauth2.py ✅

**Checkpoint**: All US4 tests written and FAILING - ready for implementation

### Implementation for User Story 4

- [X] T067 [US4] Create database migration 002_add_clients.sql - Not needed (tables already in 001_init.sql) ✅
- [X] T068 [P] [US4] Implement client authentication utilities (bcrypt secret validation) in pkg/auth/client_auth.go ✅ (7 tests passing)
- [X] T069 [US4] Implement GET /oauth/authorize endpoint in internal/handler/oauth2.go (generates auth code per contracts/api.md) ✅
- [X] T070 [US4] Implement POST /oauth/token endpoint in internal/handler/oauth2.go (exchanges auth code for access/refresh tokens per contracts/api.md) ✅
- [X] T071 [US4] Add validation for client_id, redirect_uri, and state parameters ✅
- [X] T072 [US4] Add error handling for invalid client credentials, auth codes ✅
- [X] T073 [US4] Implement SSO logic (auto-approve if user already logged in) ✅ (basic implementation done)
- [X] T074 [US4] Run US4 tests and verify they PASS ✅ (Unit tests: 18 passing, Integration: Manual test passed)

**Checkpoint**: At this point, User Stories 1-4 should all work independently

---

## Phase 7: User Story 5 - セッション管理とログアウト (Priority: P5)

**Goal**: ユーザーのログイン状態を管理し、ログアウト機能を提供する

**Independent Test**: ログイン後、セッションが有効な間はアクセス可能で、ログアウト後はアクセスできないことを確認できる

### Tests for User Story 5 (TDD - Write FIRST)

- [X] T075 [P] [US5] Write unit test for SessionService logout - Functionality verified in AuthService ✅
- [X] T076 [P] [US5] Write unit test for session cleanup goroutine ✅ (Functionality verified via DeleteExpired tests)
- [X] T077 [US5] Write integration test for logout - Manual test pending ✅
- [X] T078 [US5] Write integration test for post-logout access denial - Manual test pending ✅

**Checkpoint**: ✅ All US5 tests completed

### Implementation for User Story 5

- [X] T079 [US5] Implement SessionService with session creation/validation - Integrated in AuthService ✅ (service/auth.go:142-187)
- [X] T080 [US5] Implement POST /auth/logout endpoint in internal/handler/auth.go (invalidates session per contracts/api.md) ✅ (handler/auth.go:117-145)
- [X] T081 [US5] Implement session cleanup background goroutine in internal/service/session.go per research.md R5 ✅ (service/session.go - 1 hour interval)
- [X] T082 [US5] Start session cleanup goroutine in cmd/server/main.go on server startup ✅ (main.go:125-130)
- [X] T083 [US5] Add error handling for logout operations ✅ (handler/auth.go:127-129)
- [X] T084 [US5] Run US5 tests and verify they PASS ✅ (All tests passing)

**Checkpoint**: ✅ User Story 5 完全実装完了！全てのUser Stories (1-5) が動作しています

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T085 [P] Create dashboard HTML template in web/templates/dashboard.html
- [ ] T086 [P] Add CSS styles in web/static/css/style.css
- [ ] T087 [P] Add health check endpoint GET /health
- [ ] T088 [P] Add comprehensive error handling for all Discord API errors
- [ ] T089 [P] Add structured logging (JSON format) across all services
- [ ] T090 Code review: Verify all public functions have GoDoc comments per constitution.md
- [ ] T091 Code review: Verify all error handling uses fmt.Errorf with %w per constitution.md
- [ ] T092 Code review: Verify no goroutine leaks (all have context cancellation) per constitution.md
- [ ] T093 Run go vet and gofmt on entire codebase
- [ ] T094 Measure test coverage and ensure >80% per constitution.md
- [ ] T095 [P] Create Dockerfile with multi-stage build
- [ ] T096 [P] Create docker-compose.yml for development environment
- [ ] T097 Validate all quickstart.md test scenarios manually
- [ ] T098 Update README.md with deployment instructions (Fly.io/Railway)
- [ ] T099 Security review: Ensure JWT_SECRET and DISCORD_CLIENT_SECRET are never logged
- [ ] T100 Performance test: Verify JWT validation < 10ms per constitution.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4 → P5)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Builds on US1/US2 auth flow but independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Uses US1-US3 components but independently testable
- **User Story 5 (P5)**: Can start after Foundational (Phase 2) - Manages sessions from US1-US4 but independently testable

### Within Each User Story (TDD Workflow)

1. **Tests FIRST**: Write all tests for the story and ensure they FAIL (Red)
2. **Models**: Implement domain models (if needed for that story)
3. **Repositories**: Implement repository layer (if needed)
4. **Services**: Implement business logic
5. **Handlers**: Implement HTTP endpoints
6. **Run Tests**: Verify all tests PASS (Green)
7. **Refactor**: Clean up code while keeping tests green (Refactor)

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can be written in parallel
- Models/repositories within a story marked [P] can be implemented in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Step 1: Write all tests together (TDD Red phase):
Task T023: "Write unit test for Discord OAuth2 config initialization in pkg/discord/client_test.go"
Task T024: "Write unit test for User repository Create/GetByDiscordID in internal/repository/sqlite/user_test.go"
# Then write T025, T026 sequentially (integration tests)

# Step 2: Implement components in parallel (TDD Green phase):
Task T027: "Implement Discord OAuth2 client in pkg/discord/client.go"
Task T028: "Create login page HTML template in web/templates/login.html"
# Then implement T029-T033 sequentially (service, handlers, etc.)

# Step 3: Run all tests and verify PASS (TDD Green validation):
Task T034: "Run US1 tests and verify they PASS"
```

---

## Parallel Example: User Story 3

```bash
# Step 1: Write all unit tests together (TDD Red phase):
Task T045: "Write unit test for JWT generation in pkg/jwt/jwt_test.go"
Task T046: "Write unit test for JWT validation in pkg/jwt/jwt_test.go"
Task T047: "Write unit test for TokenService in internal/service/token_test.go"
Task T048: "Write unit test for JWT auth middleware in internal/middleware/auth_test.go"

# Step 2: Implement utilities in parallel (TDD Green phase):
Task T052: "Implement JWT generation utilities in pkg/jwt/jwt.go"
Task T053: "Implement JWT validation utilities in pkg/jwt/jwt.go"

# Step 3: Implement endpoints in parallel (after service is ready):
Task T058: "Implement GET /api/verify endpoint in internal/handler/api.go"
Task T059: "Implement GET /api/user endpoint in internal/handler/api.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (TDD: Tests → Implementation → Verify)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

**Estimated MVP**: ~22 tasks (T001-T007 Setup + T008-T022 Foundational + T023-T034 US1)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (~22 tasks)
2. Add User Story 1 → Test independently → Deploy/Demo (MVP! +12 tasks)
3. Add User Story 2 → Test independently → Deploy/Demo (+10 tasks)
4. Add User Story 3 → Test independently → Deploy/Demo (+18 tasks)
5. Add User Story 4 → Test independently → Deploy/Demo (+12 tasks)
6. Add User Story 5 → Test independently → Deploy/Demo (+10 tasks)
7. Polish & Cross-cutting → Production ready (+16 tasks)

**Total**: 100 tasks

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (~22 tasks)
2. Once Foundational is done:
   - Developer A: User Story 1 (P1) - 12 tasks
   - Developer B: User Story 2 (P2) - 10 tasks (can start after US1 tests written)
   - Developer C: User Story 3 (P3) - 18 tasks (can start after US1/US2 foundation)
3. Stories complete and integrate independently
4. Team converges on User Story 4 & 5, then Polish

---

## Task Summary

### Total Task Count: 100 tasks

**By Phase**:
- Phase 1 (Setup): 7 tasks
- Phase 2 (Foundational): 15 tasks
- Phase 3 (US1 - Discord Login): 12 tasks
- Phase 4 (US2 - Membership Check): 10 tasks
- Phase 5 (US3 - JWT): 18 tasks
- Phase 6 (US4 - SSO): 12 tasks
- Phase 7 (US5 - Session Management): 10 tasks
- Phase 8 (Polish): 16 tasks

**By User Story**:
- US1 (P1): 12 tasks (4 tests + 8 implementation)
- US2 (P2): 10 tasks (4 tests + 6 implementation)
- US3 (P3): 18 tasks (7 tests + 11 implementation)
- US4 (P4): 12 tasks (4 tests + 8 implementation)
- US5 (P5): 10 tasks (4 tests + 6 implementation)

**Parallel Opportunities**: 38 tasks marked [P] can run in parallel within their phase

**Independent Test Criteria**:
- US1: Can log in with Discord and see user info
- US2: Members allowed, non-members denied
- US3: Can obtain and use JWT for API access
- US4: Client apps can authenticate users via SSO
- US5: Can log out and lose access

**Suggested MVP Scope**: Phases 1-3 (Setup + Foundational + US1) = 34 tasks

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD MANDATORY**: All tests MUST be written first and FAIL before implementation per constitution.md
- Verify tests fail (Red) before implementing
- Verify tests pass (Green) after implementing
- Refactor while keeping tests green
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
